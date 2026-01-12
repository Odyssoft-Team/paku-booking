package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"paku-booking/internal/booking"
	"paku-booking/internal/booking/usecases"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// @Summary Create hold
// @Description Crea un hold (bloqueo temporal de cupos) para un servicio/fecha/slot.
// @Tags booking
// @Accept json
// @Produce json
// @Param body body CreateHoldRequest true "Create hold request"
// @Success 201 {object} CreateHoldResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /holds [post]
func (h *Handlers) CreateHold(w http.ResponseWriter, r *http.Request) {
	var req CreateHoldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.ServiceID == "" || req.Date == "" || req.Slot == "" {
		writeErr(w, http.StatusBadRequest, "service_id, date, slot are required")
		return
	}

	date, err := time.Parse(booking.DateLayout, req.Date)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid date (expected YYYY-MM-DD)")
		return
	}
	date = booking.NormalizeDate(date)

	slot, err := booking.ParseSlot(req.Slot)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.createHoldUC.Execute(r.Context(), usecases.CreateHoldInput{
		ServiceID:  req.ServiceID,
		LocationID: req.LocationID,
		Date:       date,
		Slot:       slot,
		Qty:        req.Qty,
	})
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}

	writeJSON(w, http.StatusCreated, CreateHoldResponse{
		HoldID:    res.HoldID,
		ExpiresAt: res.ExpiresAt.Format(time.RFC3339),
	})
}

// @Summary Cancel hold
// @Description Cancela un hold. Es idempotente; si el hold no está activo no hace nada.
// @Tags booking
// @Accept json
// @Produce json
// @Param id path string true "Hold ID"
// @Success 204 {object} nil
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /holds/{id}/cancel [post]
func (h *Handlers) CancelHold(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid hold id")
		return
	}

	err := h.cancelHoldUC.Execute(r.Context(), usecases.CancelHoldInput{HoldID: id})
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
