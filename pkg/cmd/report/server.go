package report

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type reportServer struct {
	saveTo        string
	serverAddress string
}

func newReportServer(serverAddr string, bindDir string) (*reportServer, error) {
	http.Handle("/api/v0/", apiHandler())
	http.Handle("/", handlerLoggingMiddleware(http.FileServer(http.Dir(bindDir))))

	return &reportServer{
		saveTo:        bindDir,
		serverAddress: serverAddr,
	}, nil
}

func (s *reportServer) Start() error {
	if err := http.ListenAndServe(s.serverAddress, nil); err != nil {
		return fmt.Errorf("unable to start the report server at address %s: %v", s.serverAddress, err)
	}
	return nil
}

// loggingMiddleware creates a middleware that logs HTTP requests with detailed information
func handlerLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code and response size
		wrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     200, // default to 200 if WriteHeader is not called
			responseSize:   0,
		}

		// Log the incoming request
		log.Debugf("HTTP Request: %s %s from %s - User-Agent: %s",
			r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

		// Call the next handler
		next.ServeHTTP(wrapper, r)

		// Calculate duration
		duration := time.Since(start)

		// Log the response details
		log.Debugf("HTTP Response: %s %s - Status: %d - Size: %d bytes - Duration: %v",
			r.Method, r.URL.Path, wrapper.statusCode, wrapper.responseSize, duration)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code and response size
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(size)
	return size, err
}
