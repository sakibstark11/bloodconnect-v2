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

type NotificationResponse struct {
	ID        string                  `json:"id"`
	Type      domain.NotificationType `json:"type"`
	Title     string                  `json:"title"`
	Content   string                  `json:"content"`
	Metadata  any                     `json:"metadata"`
	CreatedAt string                  `json:"created_at"`
}

func mapNotificationToResponse(n *domain.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        string(n.ID),
		Type:      n.Type,
		Title:     n.Title,
		Content:   n.Content,
		Metadata:  n.Metadata,
		CreatedAt: string(n.CreatedAt),
	}
}

type NotificationsResponse struct {
	Notifications      []NotificationResponse `json:"notifications"`
	LastNotificationID domain.NotificationID  `json:"last_notification_id,omitempty"`
	PageSize           int                    `json:"page_size"`
}

func (h *NotificationHandler) GetForMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.UserIDKey).(domain.UserID)
	if !ok || userID == "" {
		RespondWithError(w, domain.ErrUnauthorized)
		return
	}

	q := r.URL.Query()
	lastNotificationID := domain.NotificationID(q.Get("last_notification_id"))
	pageSize := h.config.DefaultPageSize

	notifications, err := h.service.GetForUser(r.Context(), userID, lastNotificationID, pageSize)
	if err != nil {
		RespondWithError(w, err)
		return
	}

	if notifications == nil {
		notifications = []*domain.Notification{}
	}

	var newLastID domain.NotificationID
	if len(notifications) > 0 {
		newLastID = notifications[len(notifications)-1].ID
	}

	res := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		res[i] = mapNotificationToResponse(n)
	}

	RespondJSON(w, http.StatusOK, NotificationsResponse{
		Notifications:      res,
		LastNotificationID: newLastID,
		PageSize:           pageSize,
	})
}
