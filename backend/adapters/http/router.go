package http

import (
	"net/http"

	"bloodconnect/adapters/http/handlers"
	"bloodconnect/application/services"
)

// SetupRouter wires all routes and applies middleware.
// /users/me/* routes are wrapped with InjectUserID middleware.
func SetupRouter(
	userService services.UserService,
	notifService services.NotificationService,
	requestService services.RequestService,
) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handlers.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	uh := handlers.NewUserHandler(userService)
	nh := handlers.NewNotificationHandler(notifService)
	rh := handlers.NewRequestHandler(requestService)

	// ── Public routes (no auth required) ───────────────────────────────────
	uh.RegisterPublicRoutes(mux)
	rh.RegisterPublicRoutes(mux)

	// ── Protected /users/me/* routes (InjectUserID middleware) ─────────────
	meMux := http.NewServeMux()
	uh.RegisterMeRoutes(meMux)
	nh.RegisterMeRoutes(meMux)
	rh.RegisterMeRoutes(meMux)

	mux.Handle("/users/me", InjectUserID(meMux))
	mux.Handle("/users/me/", InjectUserID(meMux))

	return enableCORS(mux)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
