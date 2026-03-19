package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bloodconnect/adapters/http/handlers"
	"bloodconnect/application/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			traceID := "trace_" + ulid.Make().String()

			ctx := context.WithValue(r.Context(), domain.TraceIDKey, traceID)
			r = r.WithContext(ctx)

			reqLogger := logger.With(
				zap.String("trace_id", traceID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			reqLogger.Info("Completed HTTP request",
				zap.Int("status", rw.status),
				zap.Duration("latency", duration),
			)
		})
	}
}

func GetTraceID(ctx context.Context) string {
	if val, ok := ctx.Value(domain.TraceIDKey).(string); ok {
		return val
	}
	return ""
}

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				handlers.RespondJSONError(w, http.StatusUnauthorized, "Authorization header with Bearer token is required", nil)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					handlers.RespondJSONError(w, http.StatusUnauthorized, "Invalid or expired token", nil)
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				handlers.RespondJSONError(w, http.StatusUnauthorized, "Invalid or expired token", nil)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				handlers.RespondJSONError(w, http.StatusUnauthorized, "Invalid token claims", nil)
				return
			}

			userID, ok := claims["user_id"].(string)
			if !ok || userID == "" {
				handlers.RespondJSONError(w, http.StatusUnauthorized, "User ID missing from token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), domain.UserIDKey, domain.UserID(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
