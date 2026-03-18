package handlers

import (
	"net/http"

	"bloodconnect/application/services"
	"bloodconnect/application/domain"
)

type NotificationHandler struct {
	service services.NotificationService
}

func NewNotificationHandler(service services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// RegisterMeRoutes registers /users/me/notifications (requires InjectUserID middleware)
func (h *NotificationHandler) RegisterMeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/me/notifications", h.GetForMe)
}

// NotificationsResponse wraps a list of notifications
type NotificationsResponse struct {
	Notifications []domain.Notification `json:"notifications"`
}

func (h *NotificationHandler) GetForMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(domain.UserIDKey).(string)
	if userID == "" {
		RespondJSONError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	notifications, err := h.service.GetForUser(r.Context(), userID)
	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to fetch notifications", err.Error())
		return
	}

	if notifications == nil {
		notifications = []domain.Notification{}
	}
	RespondJSON(w, http.StatusOK, NotificationsResponse{Notifications: notifications})
}
