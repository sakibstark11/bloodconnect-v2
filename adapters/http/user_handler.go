package http

import (
	"encoding/json"
	"net/http"

	"github.com/sakibalam/bloodconnect/application"
	"github.com/sakibalam/bloodconnect/domain"
)

type userHandler struct {
	service application.UserService
}

func newUserHandler(service application.UserService) *userHandler {
	return &userHandler{service: service}
}

func (h *userHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users/signup", h.Signup)
	mux.HandleFunc("POST /users/login", h.Login)
	mux.HandleFunc("PUT /users/health", h.UpdateHealth)
	mux.HandleFunc("PUT /users/location", h.UpdateLocation)
}

func (h *userHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.service.Signup(r.Context(), req.Name, req.Email, req.Password, req.Phone)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *userHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": user.ID, "token": "dummy-token-for-now"})
}

func (h *userHandler) UpdateHealth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string          `json:"user_id"`
		InfoType domain.InfoType `json:"info_type"`
		Details  string          `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateHealth(r.Context(), req.UserID, req.InfoType, req.Details); err != nil {
		http.Error(w, "Failed to update health", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *userHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string  `json:"user_id"`
		Lat    float64 `json:"lat"`
		Lng    float64 `json:"lng"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateLocation(r.Context(), req.UserID, req.Lat, req.Lng); err != nil {
		http.Error(w, "Failed to update location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
