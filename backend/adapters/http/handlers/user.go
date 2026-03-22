package handlers

import (
	"encoding/json"
	"net/http"

	"bloodconnect/application/domain"
	"bloodconnect/application/services"
)

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users/signup", h.Signup)
	mux.HandleFunc("POST /users/login", h.Login)
}

func (h *UserHandler) RegisterMeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.GetMe)
	mux.HandleFunc("PUT /health", h.UpdateHealth)
	mux.HandleFunc("PUT /location", h.UpdateLocation)
}

type SignupRequest struct {
	Name     string `json:"name"     validate:"required"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Phone    string `json:"phone"    validate:"required"`
}

type SignupResponse struct {
	ID string `json:"id"`
}

func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}
	if err := Validate.Struct(req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Validation failed", FormatValidationErrors(err))
		return
	}

	id, err := h.service.Signup(r.Context(), req.Name, req.Email, req.Password, req.Phone)
	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to create user", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, SignupResponse{ID: string(id)})
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type UserHealthResponse struct {
	InfoType string `json:"info_type"`
	Details  string `json:"details"`
}

type UserResponse struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	Email     string               `json:"email"`
	Phone     string               `json:"phone"`
	Health    []UserHealthResponse `json:"health,omitempty"`
	CreatedAt string               `json:"created_at"`
	UpdatedAt string               `json:"updated_at"`
}

func mapUserToResponse(u *domain.User, health []domain.UserHealth) UserResponse {
	healthRes := make([]UserHealthResponse, len(health))
	for i, h := range health {
		healthRes[i] = UserHealthResponse{
			InfoType: string(h.InfoType),
			Details:  h.Details,
		}
	}

	return UserResponse{
		ID:        string(u.ID),
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		Health:    healthRes,
		CreatedAt: string(u.CreatedAt),
		UpdatedAt: string(u.UpdatedAt),
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}
	if err := Validate.Struct(req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Validation failed", FormatValidationErrors(err))
		return
	}

	token, user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		RespondJSONError(w, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	RespondJSON(w, http.StatusOK, LoginResponse{ID: string(user.ID), Token: token})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, health, err := h.service.GetMe(r.Context())
	if err != nil {
		if err == domain.ErrUnauthorized {
			RespondJSONError(w, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}
		RespondJSONError(w, http.StatusInternalServerError, "Failed to get user info", err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, mapUserToResponse(user, health))
}

type UpdateHealthRequest struct {
	InfoType domain.InfoType `json:"info_type" validate:"required"`
	Details  string          `json:"details"   validate:"required"`
}

func (h *UserHandler) UpdateHealth(w http.ResponseWriter, r *http.Request) {
	var req UpdateHealthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}
	if err := Validate.Struct(req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Validation failed", FormatValidationErrors(err))
		return
	}

	if err := h.service.UpdateHealth(r.Context(), req.InfoType, req.Details); err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to update health", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdateLocationRequest struct {
	Lat float64 `json:"lat" validate:"required,latitude"`
	Lng float64 `json:"lng" validate:"required,longitude"`
}

func (h *UserHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var req UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}
	if err := Validate.Struct(req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Validation failed", FormatValidationErrors(err))
		return
	}

	if err := h.service.UpdateLocation(r.Context(), req.Lat, req.Lng); err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to update location", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
