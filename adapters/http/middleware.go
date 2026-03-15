package http

import (
	"context"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sakibalam/bloodconnect/domain"
	"go.uber.org/zap"
)

// responseWriter is a custom wrapper to capture the HTTP status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger is a middleware that injects a Trace ID and logs HTTP requests
func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			traceID := "trace_" + ulid.Make().String()
			
			// Inject traceID into context
			ctx := context.WithValue(r.Context(), domain.TraceIDKey, traceID)
			r = r.WithContext(ctx)

			// Create a scoped logger for this request
			reqLogger := logger.With(
				zap.String("trace_id", traceID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			// Wrap ResponseWriter to capture status code
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			
			// Execute actual handler
			next.ServeHTTP(rw, r)

			// Log completion
			duration := time.Since(start)
			reqLogger.Info("Completed HTTP request",
				zap.Int("status", rw.status),
				zap.Duration("latency", duration),
			)
		})
	}
}

// GetTraceID pulls the trace_id from context
func GetTraceID(ctx context.Context) string {
	if val, ok := ctx.Value(domain.TraceIDKey).(string); ok {
		return val
	}
	return ""
}
