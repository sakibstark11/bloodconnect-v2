package http

import (
	"net/http"

	"bloodconnect/adapters/http/handlers"
	"bloodconnect/application"
	"bloodconnect/application/services"
)

func SetupRouter(
	userService services.UserService,
	notifService services.NotificationService,
	requestService services.RequestService,
	config *application.AppConfig,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handlers.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	uh := handlers.NewUserHandler(userService)
	nh := handlers.NewNotificationHandler(notifService, config)
	rh := handlers.NewRequestHandler(requestService, config)

	uh.RegisterPublicRoutes(mux)
	mux.HandleFunc("GET /requests", rh.List)
	mux.HandleFunc("GET /requests/{id}", rh.Get)

	auth := AuthMiddleware(config.JWTSecret)

	mux.Handle("GET /users/me", auth(http.HandlerFunc(uh.GetMe)))
	mux.Handle("PUT /users/me/health", auth(http.HandlerFunc(uh.UpdateHealth)))
	mux.Handle("PUT /users/me/location", auth(http.HandlerFunc(uh.UpdateLocation)))

	mux.Handle("GET /notifications", auth(http.HandlerFunc(nh.GetForMe)))

	mux.Handle("POST /requests", auth(http.HandlerFunc(rh.Submit)))
	mux.Handle("POST /requests/{id}/respond", auth(http.HandlerFunc(rh.Respond)))
	mux.Handle("POST /requests/{id}/cancel", auth(http.HandlerFunc(rh.Cancel)))

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
