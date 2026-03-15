package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sakibalam/bloodconnect/application"
	"github.com/sakibalam/bloodconnect/domain"
)

type requestHandler struct {
	service application.RequestService
}

func newRequestHandler(service application.RequestService) *requestHandler {
	return &requestHandler{service: service}
}

func (h *requestHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /requests", h.Submit)
	mux.HandleFunc("POST /requests/{id}/respond", h.Respond)
	mux.HandleFunc("POST /requests/{id}/cancel", h.Cancel)
	mux.HandleFunc("GET /requests/{id}", h.Get)
}

func (h *requestHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID         string    `json:"user_id" validate:"required"`
		LocationLat    float64   `json:"location_lat" validate:"required,latitude"`
		LocationLng    float64   `json:"location_lng" validate:"required,longitude"`
		BagCount       int       `json:"bag_count" validate:"required,min=1"`
		RequiredByDate time.Time `json:"required_by_date" validate:"required"`
		BloodType      string    `json:"blood_type" validate:"required"`
		ContactPhone   string    `json:"contact_phone" validate:"required"`
		Description    string    `json:"description"`
		RequesterInfo  string    `json:"requester_info"`
		LocationName   string    `json:"location_name" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	if err := Validate.Struct(req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Validation failed", FormatValidationErrors(err))
		return
	}

	id, err := h.service.SubmitRequest(r.Context(),
		req.UserID, req.BloodType, req.ContactPhone, req.Description,
		req.RequesterInfo, req.LocationName, req.LocationLat, req.LocationLng, req.BagCount, req.RequiredByDate)

	if err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to submit request", err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *requestHandler) Respond(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("id")
	if requestID == "" {
		RespondJSONError(w, http.StatusBadRequest, "Missing request ID in URL", nil)
		return
	}

	var req struct {
		UserID string              `json:"user_id" validate:"required"`
		Action domain.ActionStatus `json:"action" validate:"required,oneof=Accepted Declined Donated"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	if err := Validate.Struct(req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Validation failed", FormatValidationErrors(err))
		return
	}

	if err := h.service.RespondToRequest(r.Context(), requestID, req.UserID, req.Action); err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to respond to request", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *requestHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("id")
	if requestID == "" {
		RespondJSONError(w, http.StatusBadRequest, "Missing request ID in URL", nil)
		return
	}

	var req struct {
		UserID string `json:"user_id" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	if err := Validate.Struct(req); err != nil {
		RespondJSONError(w, http.StatusBadRequest, "Validation failed", FormatValidationErrors(err))
		return
	}

	if err := h.service.CancelRequest(r.Context(), requestID, req.UserID); err != nil {
		RespondJSONError(w, http.StatusInternalServerError, "Failed to cancel request", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *requestHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
}
