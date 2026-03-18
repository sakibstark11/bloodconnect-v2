package handlers

import (
	"net/http"
	"strconv"

	"bloodconnect/application"
	"bloodconnect/application/domain"
	"bloodconnect/application/services"
)

type RequestHandler struct {
	service services.RequestService
}

func NewRequestHandler(service services.RequestService) *RequestHandler {
	return &RequestHandler{service: service}
}

// RegisterPublicRoutes registers routes that don't require authentication
func (h *RequestHandler) RegisterPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /requests", h.Submit)
	mux.HandleFunc("GET /requests", h.List)
	mux.HandleFunc("GET /requests/{id}", h.Get)
}

// RegisterMeRoutes registers /users/me/requests/* routes
func (h *RequestHandler) RegisterMeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users/me/requests/{id}/respond", h.Respond)
	mux.HandleFunc("POST /users/me/requests/{id}/cancel", h.Cancel)
}

// SubmitRequestBody is the expected body for POST /requests
type SubmitRequestBody struct {
	UserID         string  `json:"user_id"          validate:"required"`
	LocationLat    float64 `json:"location_lat"     validate:"required,latitude"`
	LocationLng    float64 `json:"location_lng"     validate:"required,longitude"`
	BagCount       int     `json:"bag_count"        validate:"required,min=1"`
	RequiredByDate string  `json:"required_by_date" validate:"required"`
	BloodType      string  `json:"blood_type"       validate:"required"`
	ContactPhone   string  `json:"contact_phone"    validate:"required"`
	Description    string  `json:"description"`
	RequesterInfo  string  `json:"requester_info"`
	LocationName   string  `json:"location_name"    validate:"required"`
}

// SubmitRequestResponse is the response body for POST /requests
type SubmitRequestResponse struct {
	ID string `json:"id"`
}

func (h *RequestHandler) Submit(w http.ResponseWriter, r *http.Request) {
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
		RespondJSONError(w, http.StatusBadRequest, "Invalid required_by_date format (use RFC3339)", err.Error())
		return
	}

	id, err := h.service.SubmitRequest(r.Context(),
		req.UserID, req.BloodType, req.ContactPhone, req.Description,
		req.RequesterInfo, req.LocationName, req.LocationLat, req.LocationLng,
		req.BagCount, requiredBy)
	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to submit request", err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, SubmitRequestResponse{ID: id})
}

// ListRequestsResponse wraps a paginated list of requests
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
	// page_size is system-defined, not user-defined
	pageSize := 20

	filters := application.RequestFilters{
		BloodType:   q.Get("blood_type"),
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
	requestID := r.PathValue("id")
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

// RespondToRequestBody is the expected body for POST /users/me/requests/{id}/respond
type RespondToRequestBody struct {
	Action domain.ActionStatus `json:"action" validate:"required,oneof=Accepted Declined Donated"`
}

func (h *RequestHandler) Respond(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("id")
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
	requestID := r.PathValue("id")
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
