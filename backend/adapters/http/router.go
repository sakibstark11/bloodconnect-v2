package http

import (
	"encoding/json"
	"net/http"

	"github.com/sakibalam/bloodconnect/application"
)

// SetupRouter creates the base ServeMux for the application and wraps it with CORS
func SetupRouter(
	userService application.UserService,
	notifService application.NotificationService,
	requestService application.RequestService,
) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// User routes
	uh := newUserHandler(userService)
	uh.RegisterRoutes(mux)

	// Notification routes
	nh := newNotificationHandler(notifService)
	nh.RegisterRoutes(mux)

	// Request routes
	rh := newRequestHandler(requestService)
	rh.RegisterRoutes(mux)

	return enableCORS(mux)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins for development
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
