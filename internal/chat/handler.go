package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/vertex"
	log "github.com/sirupsen/logrus"
)

const defaultModel = "claude-sonnet-4-5"

type Config struct {
	ReportDir string
	Enabled   bool
	Provider  string // "vertex", "anthropic", or ""
	Model     string
}

type Handler struct {
	config   Config
	client   *anthropic.Client
	executor *ToolExecutor
	prompt   string
}

func DetectConfig(reportDir string) Config {
	cfg := Config{
		ReportDir: reportDir,
		Model:     defaultModel,
	}

	if os.Getenv("CLAUDE_CODE_USE_VERTEX") == "1" {
		region := os.Getenv("GOOGLE_CLOUD_LOCATION")
		if region == "" {
			region = os.Getenv("CLOUD_ML_REGION")
		}
		projectID := os.Getenv("ANTHROPIC_VERTEX_PROJECT_ID")
		if projectID == "" {
			projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
		}
		if region != "" && projectID != "" {
			cfg.Enabled = true
			cfg.Provider = "vertex"
			return cfg
		}
	}

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		cfg.Enabled = true
		cfg.Provider = "anthropic"
		return cfg
	}

	return cfg
}

func NewHandler(cfg Config) *Handler {
	h := &Handler{
		config:   cfg,
		executor: NewToolExecutor(cfg.ReportDir),
		prompt:   LoadSystemPrompt(cfg.ReportDir),
	}

	if !cfg.Enabled {
		return h
	}

	switch cfg.Provider {
	case "vertex":
		region := os.Getenv("GOOGLE_CLOUD_LOCATION")
		if region == "" {
			region = os.Getenv("CLOUD_ML_REGION")
		}
		projectID := os.Getenv("ANTHROPIC_VERTEX_PROJECT_ID")
		if projectID == "" {
			projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
		}
		client := anthropic.NewClient(
			vertex.WithGoogleAuth(context.Background(), region, projectID),
		)
		h.client = &client
	case "anthropic":
		client := anthropic.NewClient(
			option.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		)
		h.client = &client
	}

	log.Infof("Chat assistant enabled (provider: %s, model: %s)", cfg.Provider, cfg.Model)
	return h
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/chat/status", h.handleStatus)
	mux.HandleFunc("/api/v1/chat", h.handleChat)
	mux.HandleFunc("/api/v1/chat/sessions", h.handleSessions)
	mux.HandleFunc("/api/v1/chat/sessions/", h.handleSessionByID)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":  h.config.Enabled,
		"provider": h.config.Provider,
		"model":    h.config.Model,
	})
}

type ChatRequest struct {
	Message string        `json:"message"`
	History []ChatMessage `json:"history"`
}

func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.config.Enabled || h.client == nil {
		http.Error(w, "chat not enabled", http.StatusServiceUnavailable)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	messages := buildMessages(req.History, req.Message)
	tools := ToolDefinitions()
	systemPrompt := []anthropic.TextBlockParam{{Text: h.prompt}}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		stream := h.client.Messages.NewStreaming(r.Context(), anthropic.MessageNewParams{
			Model:     h.config.Model,
			MaxTokens: 8192,
			System:    systemPrompt,
			Messages:  messages,
			Tools:     tools,
		})

		msg := anthropic.Message{}
		var currentText strings.Builder

		for stream.Next() {
			event := stream.Current()
			if err := msg.Accumulate(event); err != nil {
				log.Errorf("stream accumulate error: %v", err)
				break
			}

			switch ev := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch delta := ev.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					currentText.WriteString(delta.Text)
					sendSSE(w, flusher, "text", delta.Text)
				}
			}
		}

		if stream.Err() != nil {
			log.Errorf("stream error: %v", stream.Err())
			sendSSE(w, flusher, "error", stream.Err().Error())
			return
		}

		if msg.StopReason == "tool_use" {
			messages = append(messages, msg.ToParam())
			toolResults := []anthropic.ContentBlockParamUnion{}

			for _, block := range msg.Content {
				if tu, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
					inputJSON, _ := json.Marshal(tu.Input)
					log.Debugf("Tool call: %s(%s)", tu.Name, string(inputJSON))
					sendSSE(w, flusher, "tool_call", fmt.Sprintf(`{"name":"%s"}`, tu.Name))

					result, err := h.executor.Execute(tu.Name, inputJSON)
					if err != nil {
						log.Warnf("Tool execution error %s: %v", tu.Name, err)
						toolResults = append(toolResults, anthropic.NewToolResultBlock(tu.ID, fmt.Sprintf("Error: %v", err), true))
					} else {
						toolResults = append(toolResults, anthropic.NewToolResultBlock(tu.ID, result, false))
					}
				}
			}

			messages = append(messages, anthropic.NewUserMessage(toolResults...))
			continue
		}

		sendSSE(w, flusher, "done", currentText.String())
		return
	}
}

func buildMessages(history []ChatMessage, newMessage string) []anthropic.MessageParam {
	var messages []anthropic.MessageParam
	for _, msg := range history {
		switch msg.Role {
		case "user":
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case "assistant":
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(newMessage)))
	return messages
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, event, data string) {
	escaped := strings.ReplaceAll(data, "\n", "\\n")
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, escaped)
	flusher.Flush()
}

func (h *Handler) handleSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		sessions, err := ListSessions(h.config.ReportDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(sessions)

	case http.MethodPost:
		var session Session
		if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if session.ID == "" {
			s := NewSession()
			session.ID = s.ID
			session.Created = s.Created
		}
		if err := SaveSession(h.config.ReportDir, &session); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"id": session.ID})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/chat/sessions/")
	if id == "" {
		http.Error(w, "missing session id", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		session, err := LoadSession(h.config.ReportDir, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(session)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
