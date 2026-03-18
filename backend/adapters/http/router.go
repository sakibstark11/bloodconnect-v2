package http

import (
	"net/http"

	"bloodconnect/adapters/http/handlers"
	"bloodconnect/application"
	"bloodconnect/application/services"
)

// SetupRouter wires all routes and applies middleware.
// /users/me/* routes are wrapped with AuthMiddleware.
func SetupRouter(
	userService services.UserService,
	notifService services.NotificationService,
	requestService services.RequestService,
	config *application.AppConfig,
) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handlers.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	uh := handlers.NewUserHandler(userService)
	nh := handlers.NewNotificationHandler(notifService, config)
	rh := handlers.NewRequestHandler(requestService, config)

	// ── Public routes ──────────────────────────────────────────────────────
	uh.RegisterPublicRoutes(mux)
	rh.RegisterPublicRoutes(mux) // POST /requests, GET /requests, GET /requests/{id}

	// ── Protected routes (AuthMiddleware with JWT verification) ───────────
	auth := AuthMiddleware(config.JWTSecret)

	// Users /me
	meMux := http.NewServeMux()
	uh.RegisterMeRoutes(meMux)
	mux.Handle("/users/me", auth(http.StripPrefix("/users/me", meMux)))
	mux.Handle("/users/me/", auth(http.StripPrefix("/users/me", meMux)))

	// Notifications
	notifMux := http.NewServeMux()
	nh.RegisterRoutes(notifMux)
	mux.Handle("/notifications", auth(http.StripPrefix("/notifications", notifMux)))
	mux.Handle("/notifications/", auth(http.StripPrefix("/notifications", notifMux)))

	// Requests (Protected actions)
	reqMux := http.NewServeMux()
	rh.RegisterProtectedRoutes(reqMux)
	mux.Handle("/requests/", auth(http.StripPrefix("/requests", reqMux)))

	return enableCORS(mux)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
