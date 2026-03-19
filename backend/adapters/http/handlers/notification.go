package handlers

import (
	"net/http"

	"bloodconnect/application"
	"bloodconnect/application/domain"
	"bloodconnect/application/services"
)

type NotificationHandler struct {
	service services.NotificationService
	config  *application.AppConfig
}

func NewNotificationHandler(service services.NotificationService, config *application.AppConfig) *NotificationHandler {
	return &NotificationHandler{service: service, config: config}
}

func (h *NotificationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.GetForMe)
}

type NotificationsResponse struct {
	Notifications      []domain.NotificationForUser `json:"notifications"`
	LastNotificationID domain.NotificationID        `json:"last_notification_id,omitempty"`
	PageSize           int                          `json:"page_size"`
}

func (h *NotificationHandler) GetForMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.UserIDKey).(domain.UserID)
	if !ok || userID == "" {
		RespondJSONError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	q := r.URL.Query()
	lastNotificationID := domain.NotificationID(q.Get("last_notification_id"))
	pageSize := h.config.DefaultPageSize

	notifications, err := h.service.GetForUser(r.Context(), userID, lastNotificationID, pageSize)
	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to fetch notifications", err.Error())
		return
	}

	if notifications == nil {
		notifications = []domain.NotificationForUser{}
	}

	var newLastID domain.NotificationID
	if len(notifications) > 0 {
		newLastID = notifications[len(notifications)-1].ID
	}

	RespondJSON(w, http.StatusOK, NotificationsResponse{
		Notifications:      notifications,
		LastNotificationID: newLastID,
		PageSize:           pageSize,
	})
}
