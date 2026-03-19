package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bloodconnect/application"
	"bloodconnect/application/domain"
	"bloodconnect/application/services"
)

type RequestHandler struct {
	service services.RequestService
	config  *application.AppConfig
}

func NewRequestHandler(service services.RequestService, config *application.AppConfig) *RequestHandler {
	return &RequestHandler{service: service, config: config}
}

func (h *RequestHandler) RegisterPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /requests", h.List)
	mux.HandleFunc("GET /requests/{id}", h.Get)
}

func (h *RequestHandler) RegisterProtectedRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /", h.Submit)
	mux.HandleFunc("POST /{id}/respond", h.Respond)
	mux.HandleFunc("POST /{id}/cancel", h.Cancel)
}

type SubmitRequestBody struct {
	LocationLat    float64 `json:"location_lat"     validate:"required,latitude"`
	LocationLng    float64 `json:"location_lng"     validate:"required,longitude"`
	BagCount       int     `json:"bag_count"        validate:"required,min=1"`
	RequiredByDate string  `json:"required_by_date" validate:"required"`
	BloodType      string  `json:"blood_type"       validate:"required"`
	Description    string  `json:"description"`
	RequesterInfo  string  `json:"requester_info"`
	LocationName   string  `json:"location_name"    validate:"required"`
}

func (h *RequestHandler) Submit(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.UserIDKey).(domain.UserID)
	if !ok || userID == "" {
		RespondJSONError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req SubmitRequestBody
	if err := decodeJSONBody(r, &req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}
	if err := Validate.Struct(req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Validation failed", FormatValidationErrors(err))
		return
	}

	requiredBy, err := parseDateTime(req.RequiredByDate)
	if err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid required_by_date format (expected ISO 8601)", err.Error())
		return
	}

	minDate := time.Now().Add(time.Duration(h.config.ProcessRequestWindowDays) * 24 * time.Hour)
	if requiredBy.Before(minDate) {
		RespondJSONError(w, http.StatusBadRequest, fmt.Sprintf("required_by_date must be at least %d days in the future", h.config.ProcessRequestWindowDays), nil)
		return
	}

	id, err := h.service.SubmitRequest(r.Context(),
		userID, domain.BloodType(req.BloodType), req.Description,
		req.RequesterInfo, req.LocationName, req.LocationLat, req.LocationLng,
		req.BagCount, domain.ISOTimestamp(req.RequiredByDate))
	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to submit request", err.Error())
		return
	}

	RespondJSON(w, http.StatusAccepted, map[string]string{"id": string(id)})
}

type ListRequestsResponse struct {
	Requests []domain.DonationRequest `json:"requests"`
	Total    int                      `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

func (h *RequestHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize := 20

	filters := application.RequestFilters{
		BloodType:   domain.BloodType(q.Get("blood_type")),
		Status:      q.Get("status"),
		LocationHex: q.Get("location_hex"),
	}

	requests, total, err := h.service.ListRequests(r.Context(), filters, page, pageSize)
	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to list requests", err.Error())
		return
	}
	if requests == nil {
		requests = []domain.DonationRequest{}
	}

	RespondJSON(w, http.StatusOK, ListRequestsResponse{
		Requests: requests,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *RequestHandler) Get(w http.ResponseWriter, r *http.Request) {
	requestID := domain.RequestID(r.PathValue("id"))
	if requestID == "" {
		RespondJSONError(w, http.StatusBadRequest, "Missing request ID in URL", nil)
		return
	}

	req, err := h.service.GetExtendedRequest(r.Context(), requestID)
	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to get request", err.Error())
		return
	}
	if req == nil {
		RespondJSONError(w, http.StatusNotFound, "Request not found", nil)
		return
	}

	RespondJSON(w, http.StatusOK, req)
}

type RespondToRequestBody struct {
	Action domain.ActionStatus `json:"action" validate:"required,oneof=Accepted Declined Donated"`
}

func (h *RequestHandler) Respond(w http.ResponseWriter, r *http.Request) {
	requestID := domain.RequestID(r.PathValue("id"))
	if requestID == "" {
		RespondJSONError(w, http.StatusBadRequest, "Missing request ID in URL", nil)
		return
	}

	var req RespondToRequestBody
	if err := decodeJSONBody(r, &req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}
	if err := Validate.Struct(req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Validation failed", FormatValidationErrors(err))
		return
	}

	if err := h.service.RespondToRequest(r.Context(), requestID, req.Action); err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to respond to request", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RequestHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	requestID := domain.RequestID(r.PathValue("id"))
	if requestID == "" {
		RespondJSONError(w, http.StatusBadRequest, "Missing request ID in URL", nil)
		return
	}

	if err := h.service.CancelRequest(r.Context(), requestID); err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to cancel request", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
