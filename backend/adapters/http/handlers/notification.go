package handlers

import (
	"net/http"

	"bloodconnect/application"
	"bloodconnect/application/domain"
	"bloodconnect/application/services"
	"strconv"
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
	Notifications []domain.Notification `json:"notifications"`
	Total         int                   `json:"total"`
	Page          int                   `json:"page"`
	PageSize      int                   `json:"page_size"`
}

func (h *NotificationHandler) GetForMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.UserIDKey).(domain.UserID)
	if !ok || userID == "" {
		RespondJSONError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := h.config.NotificationPageSize

	notifications, total, err := h.service.GetForUser(r.Context(), userID, page, pageSize)
	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to fetch notifications", err.Error())
		return
	}

	if notifications == nil {
		notifications = []domain.Notification{}
	}
	RespondJSON(w, http.StatusOK, NotificationsResponse{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	})
}
