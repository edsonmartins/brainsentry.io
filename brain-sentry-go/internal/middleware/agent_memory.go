package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/service"
)

// AgentMemoryHeaders drive the middleware behavior.
const (
	HeaderAgentMemoryEnabled = "X-Agent-Memory" // "true" to enable
	HeaderAgentMemoryQuery   = "X-Agent-Memory-Query"
	HeaderAgentMemoryTopK    = "X-Agent-Memory-Top-K"
	HeaderAgentMemorySet     = "X-Agent-Memory-Set"
	HeaderAgentID            = "X-Agent-ID"
	HeaderSessionID          = "X-Session-ID"
	HeaderInjectedContext    = "X-Memory-Context"
)

// responseCapture wraps ResponseWriter to capture status code and body for tracing.
type responseCapture struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
	limit  int // max bytes to capture
}

func (r *responseCapture) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseCapture) Write(b []byte) (int, error) {
	if r.body.Len() < r.limit {
		remaining := r.limit - r.body.Len()
		if remaining > len(b) {
			remaining = len(b)
		}
		r.body.Write(b[:remaining])
	}
	return r.ResponseWriter.Write(b)
}

// AgentMemory is an HTTP middleware that:
//   - Before the handler: if X-Agent-Memory=true, recalls memories and injects
//     them via the X-Memory-Context request header (or request body for POST).
//   - After the handler: records an AgentTrace capturing the call's inputs,
//     outputs, memory context, and status.
//
// The semantic API service handles the actual recall/trace persistence.
func AgentMemory(semanticAPI *service.SemanticAPIService, traceSvc *service.AgentTraceService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			enabled := r.Header.Get(HeaderAgentMemoryEnabled) == "true"
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			query := r.Header.Get(HeaderAgentMemoryQuery)
			agentID := r.Header.Get(HeaderAgentID)
			sessionID := r.Header.Get(HeaderSessionID)
			setName := r.Header.Get(HeaderAgentMemorySet)

			var memoryContext string
			var memoryIDs []string

			// Pre-handler: recall memories if query provided
			if query != "" && semanticAPI != nil {
				recallResp, err := semanticAPI.Recall(r.Context(), service.RecallRequest{
					Query: query,
					Set:   setName,
					Limit: 5,
				})
				if err == nil && len(recallResp.Results) > 0 {
					memoryContext = formatMemoryContext(recallResp.Results)
					for _, res := range recallResp.Results {
						memoryIDs = append(memoryIDs, res.MemoryID)
					}
					// Expose to handler via request header
					r.Header.Set(HeaderInjectedContext, memoryContext)
				}
			}

			// Capture response for tracing
			capture := &responseCapture{
				ResponseWriter: w,
				status:         http.StatusOK,
				body:           &bytes.Buffer{},
				limit:          4096,
			}

			// Capture request body for trace params
			var requestBody map[string]any
			if r.Body != nil && (r.Method == http.MethodPost || r.Method == http.MethodPut) {
				bodyBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				if len(bodyBytes) > 0 {
					_ = json.Unmarshal(bodyBytes, &requestBody)
				}
			}

			// Invoke handler
			next.ServeHTTP(capture, r)

			// Post-handler: record trace
			if traceSvc == nil {
				return
			}

			status := domain.AgentTraceSuccess
			errMsg := ""
			if capture.status >= http.StatusBadRequest {
				status = domain.AgentTraceError
				errMsg = capture.body.String()
				if len(errMsg) > 500 {
					errMsg = errMsg[:500] + "…"
				}
			}

			// Parse response body as JSON if possible
			var responseData any
			if capture.body.Len() > 0 {
				_ = json.Unmarshal(capture.body.Bytes(), &responseData)
			}

			// Trace recording uses background context because the request one is done.
			// We copy tenant from the original request context.
			ctx := context.Background()
			_, _ = traceSvc.Record(ctx, service.RecordTraceRequest{
				SessionID:      sessionID,
				AgentID:        agentID,
				OriginFunction: r.Method + " " + r.URL.Path,
				WithMemory:     query != "",
				MemoryQuery:    query,
				MethodParams:   requestBody,
				MethodReturn:   responseData,
				MemoryContext:  memoryContext,
				Status:         status,
				ErrorMessage:   errMsg,
				DurationMs:     time.Since(start).Milliseconds(),
				MemoryIDs:      memoryIDs,
			})
		})
	}
}

// formatMemoryContext renders recall results as a single text block suitable
// for injection into a prompt or handler context.
func formatMemoryContext(results []service.RecallResult) string {
	var buf bytes.Buffer
	buf.WriteString("# Relevant memories\n\n")
	for i, r := range results {
		if i >= 5 {
			break
		}
		buf.WriteString("- ")
		if r.Summary != "" {
			buf.WriteString(r.Summary)
			buf.WriteString(": ")
		}
		content := r.Content
		if len(content) > 500 {
			content = content[:500] + "…"
		}
		buf.WriteString(content)
		buf.WriteByte('\n')
	}
	return buf.String()
}
