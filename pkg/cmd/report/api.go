package report

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

// apiHandler creates a handler for all API endpoints
func apiHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set JSON content type for all API responses
		w.Header().Set("Content-Type", "application/json")

		// Remove the /api/v0 prefix from the path
		path := strings.TrimPrefix(r.URL.Path, "/api/v0")

		switch {
		case path == "/tools":
			handleAPITools(w, r)
		case path == "/report":
			handleAPIReport(w, r)
		case path == "/report/summary":
			handleAPIReportSummary(w, r)
		case path == "/report/jobs":
			handleAPIReportJobs(w, r)
		case regexp.MustCompile(`^/report/job/[^/]+/tests$`).MatchString(path):
			handleAPIJobTests(w, r, path)
		case regexp.MustCompile(`^/report/job/[^/]+/failures$`).MatchString(path):
			handleAPIJobFailures(w, r, path)
		case regexp.MustCompile(`^/report/job/[^/]+/test/[^/]+/failure$`).MatchString(path):
			handleAPITestFailure(w, r, path)
		default:
			http.Error(w, `{"error":"endpoint not found"}`, http.StatusNotFound)
		}
	})
}

// handleAPITools returns available tools information
func handleAPITools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"tools": []map[string]string{
			{"name": "opct", "version": "v0.5.0", "description": "OpenShift Provider Certification Tool"},
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleAPIReport returns general report information
func handleAPIReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"message": "Report API endpoint - use /summary, /jobs for specific data",
		"endpoints": []string{
			"/api/v0/report/summary",
			"/api/v0/report/jobs",
			"/api/v0/report/job/<job_name>/tests",
			"/api/v0/report/job/<job_name>/failures",
			"/api/v0/report/job/<job_name>/test/<test_name>/failure",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleAPIReportSummary returns report summary
func handleAPIReportSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement actual report summary data
	response := map[string]interface{}{
		"summary": "Report summary data would be here",
		"status":  "not_implemented",
	}

	json.NewEncoder(w).Encode(response)
}

// handleAPIReportJobs returns list of jobs
func handleAPIReportJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement actual jobs data
	response := map[string]interface{}{
		"jobs": []string{
			"kubernetes-conformance",
			"openshift-conformance",
			"openshift-upgrade",
		},
		"status": "not_implemented",
	}

	json.NewEncoder(w).Encode(response)
}

// handleAPIJobTests returns tests for a specific job
func handleAPIJobTests(w http.ResponseWriter, r *http.Request, path string) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract job name from path: /report/job/{job_name}/tests
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
		return
	}
	jobName := parts[3]

	// TODO: Implement actual job tests data
	response := map[string]interface{}{
		"job":    jobName,
		"tests":  []string{},
		"status": "not_implemented",
	}

	json.NewEncoder(w).Encode(response)
}

// handleAPIJobFailures returns failures for a specific job
func handleAPIJobFailures(w http.ResponseWriter, r *http.Request, path string) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract job name from path: /report/job/{job_name}/failures
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
		return
	}
	jobName := parts[3]

	// TODO: Implement actual job failures data
	response := map[string]interface{}{
		"job":      jobName,
		"failures": []string{},
		"status":   "not_implemented",
	}

	json.NewEncoder(w).Encode(response)
}

// handleAPITestFailure returns failure details for a specific test
func handleAPITestFailure(w http.ResponseWriter, r *http.Request, path string) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract job name and test name from path: /report/job/{job_name}/test/{test_name}/failure
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
		return
	}
	jobName := parts[3]
	testName := parts[5]

	// TODO: Implement actual test failure data
	response := map[string]interface{}{
		"job":     jobName,
		"test":    testName,
		"failure": map[string]string{},
		"status":  "not_implemented",
	}

	json.NewEncoder(w).Encode(response)
}
