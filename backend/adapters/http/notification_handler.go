package http

import (
	"encoding/json"
	"net/http"

	"github.com/sakibalam/bloodconnect/application"
	"github.com/sakibalam/bloodconnect/domain"
)

type notificationHandler struct {
	service application.NotificationService
}

func newNotificationHandler(service application.NotificationService) *notificationHandler {
	return &notificationHandler{service: service}
}

func (h *notificationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /notifications", h.Submit)
	mux.HandleFunc("PUT /notifications/{id}", h.Update)
	mux.HandleFunc("DELETE /notifications/{id}", h.Delete)
	mux.HandleFunc("GET /notifications/user/{userID}", h.GetForUser)
}

func (h *notificationHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type      domain.NotificationType `json:"type"`
		Recipient string                  `json:"recipient"`
		Title     string                  `json:"title"`
		Content   string                  `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.service.Submit(r.Context(), req.Type, req.Recipient, req.Title, req.Content)
	if err != nil {
		http.Error(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *notificationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var req struct {
		Status domain.NotificationStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, "Failed to update notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *notificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *notificationHandler) GetForUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if userID == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}

	notifications, err := h.service.GetForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notifications)
}
