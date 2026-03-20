package handlers

import (
	"fmt"
	"net/http"
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

type DonationRequestResponse struct {
	ID             string              `json:"id"`
	UserID         string              `json:"user_id"`
	LocationHex    string              `json:"location_hex"`
	LocationLat    float64             `json:"location_lat"`
	LocationLng    float64             `json:"location_lng"`
	BagCount       int                 `json:"bag_count"`
	RequiredByDate string              `json:"required_by_date"`
	BloodType      domain.BloodType    `json:"blood_type"`
	Description    string              `json:"description"`
	RequesterInfo  string              `json:"requester_info"`
	LocationName   string              `json:"location_name"`
	Status         domain.RequestStatus `json:"status"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
}

func mapDonationRequestToResponse(req domain.DonationRequest) DonationRequestResponse {
	return DonationRequestResponse{
		ID:             string(req.ID),
		UserID:         string(req.UserID),
		LocationHex:    req.LocationHex,
		LocationLat:    req.LocationLat,
		LocationLng:    req.LocationLng,
		BagCount:       req.BagCount,
		RequiredByDate: string(req.RequiredByDate),
		BloodType:      req.BloodType,
		Description:    req.Description,
		RequesterInfo:  req.RequesterInfo,
		LocationName:   req.LocationName,
		Status:         req.Status,
		CreatedAt:      string(req.CreatedAt),
		UpdatedAt:      string(req.UpdatedAt),
	}
}

type RequestActionedUserResponse struct {
	UserID string              `json:"user_id"`
	Lat    float64             `json:"lat"`
	Lng    float64             `json:"lng"`
	H3Hex  string              `json:"h3_hex"`
	Action domain.ActionStatus `json:"action"`
}

func mapRequestActionedUserToResponse(u domain.RequestActionedUser) RequestActionedUserResponse {
	return RequestActionedUserResponse{
		UserID: string(u.UserID),
		Lat:    u.Lat,
		Lng:    u.Lng,
		H3Hex:  u.H3Hex,
		Action: u.Action,
	}
}

type ExtendedDonationRequestResponse struct {
	Request       DonationRequestResponse       `json:"request"`
	NotifiedUsers []RequestActionedUserResponse `json:"notified_users"`
}

func mapExtendedDonationRequestToResponse(req *domain.ExtendedDonationRequest) ExtendedDonationRequestResponse {
	notifiedUsers := make([]RequestActionedUserResponse, len(req.NotifiedUsers))
	for i, u := range req.NotifiedUsers {
		notifiedUsers[i] = mapRequestActionedUserToResponse(u)
	}
	return ExtendedDonationRequestResponse{
		Request:       mapDonationRequestToResponse(*req.Request),
		NotifiedUsers: notifiedUsers,
	}
}

type ListRequestsResponse struct {
	Requests      []DonationRequestResponse `json:"requests"`
	LastRequestID domain.RequestID          `json:"last_request_id,omitempty"`
	PageSize      int                       `json:"page_size"`
}

func (h *RequestHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	lastRequestID := domain.RequestID(q.Get("last_request_id"))
	pageSize := h.config.DefaultPageSize

	filters := application.RequestFilters{
		BloodType:   domain.BloodType(q.Get("blood_type")),
		Status:      q.Get("status"),
		LocationHex: q.Get("location_hex"),
	}

	requests, err := h.service.ListRequests(r.Context(), filters, lastRequestID, pageSize)
	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to list requests", err.Error())
		return
	}
	if requests == nil {
		requests = []domain.DonationRequest{}
	}

	var newLastID domain.RequestID
	if len(requests) > 0 {
		newLastID = requests[len(requests)-1].ID
	}

	res := make([]DonationRequestResponse, len(requests))
	for i, r := range requests {
		res[i] = mapDonationRequestToResponse(r)
	}

	RespondJSON(w, http.StatusOK, ListRequestsResponse{
		Requests:      res,
		LastRequestID: newLastID,
		PageSize:      pageSize,
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

	RespondJSON(w, http.StatusOK, mapExtendedDonationRequestToResponse(req))
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
