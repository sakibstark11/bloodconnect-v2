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
		UserID         string    `json:"user_id"`
		LocationHex    string    `json:"location_hex"`
		LocationLat    float64   `json:"location_lat"`
		LocationLng    float64   `json:"location_lng"`
		BagCount       int       `json:"bag_count"`
		RequiredByDate time.Time `json:"required_by_date"`
		BloodType      string    `json:"blood_type"`
		ContactPhone   string    `json:"contact_phone"`
		Description    string    `json:"description"`
		RequesterInfo  string    `json:"requester_info"`
		LocationName   string    `json:"location_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.service.SubmitRequest(r.Context(),
		req.UserID, req.BloodType, req.ContactPhone, req.Description,
		req.RequesterInfo, req.LocationName, req.LocationLat, req.LocationLng, req.BagCount, req.RequiredByDate)

	if err != nil {
		http.Error(w, "Failed to submit request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *requestHandler) Respond(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("id")
	if requestID == "" {
		http.Error(w, "missing request id", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID string              `json:"user_id"`
		Action domain.ActionStatus `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.RespondToRequest(r.Context(), requestID, req.UserID, req.Action); err != nil {
		http.Error(w, "Failed to respond to request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *requestHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("id")
	if requestID == "" {
		http.Error(w, "missing request id", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CancelRequest(r.Context(), requestID, req.UserID); err != nil {
		http.Error(w, "Failed to cancel request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *requestHandler) Get(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("id")
	if requestID == "" {
		http.Error(w, "missing request id", http.StatusBadRequest)
		return
	}

	req, err := h.service.GetRequest(r.Context(), requestID)
	if err != nil {
		http.Error(w, "Failed to get request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if req == nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
}
