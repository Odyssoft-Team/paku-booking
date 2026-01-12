package httpapi

import (
	"encoding/json"
	"net/http"

	"paku-booking/internal/booking"
	"paku-booking/internal/booking/usecases"

	"github.com/google/uuid"
)

// @Summary Confirm booking
// @Description Confirma un booking a partir de un hold y payment_id. Idempotente.
// @Tags booking
// @Accept json
// @Produce json
// @Param body body ConfirmBookingRequest true "Confirm booking request"
// @Success 200 {object} ConfirmBookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/confirm [post]
func (h *Handlers) ConfirmBooking(w http.ResponseWriter, r *http.Request) {
	var req ConfirmBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.HoldID == "" || req.PaymentID == "" {
		writeErr(w, http.StatusBadRequest, "hold_id and payment_id are required")
		return
	}
	if _, err := uuid.Parse(req.HoldID); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid hold_id")
		return
	}

	res, err := h.confirmBookingUC.Execute(r.Context(), usecases.ConfirmBookingInput{
		HoldID:    req.HoldID,
		PaymentID: req.PaymentID,
	})
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}

	writeJSON(w, http.StatusOK, ConfirmBookingResponse{
		BookingID: res.BookingID,
		Status:    string(booking.BookingConfirmed),
	})
}
