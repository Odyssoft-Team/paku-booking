package httpapi

import (
	"net/http"

	"paku-booking/internal/booking/usecases"
)

// @Summary List pending outbox
// @Description Devuelve mensajes pendientes del outbox (ordenados por created_at).
// @Tags booking
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" default(50)
// @Success 200 {array} booking.OutboxMessage
// @Failure 500 {object} ErrorResponse
// @Router /internal/outbox/pending [get]
func (h *Handlers) GetOutboxPending(w http.ResponseWriter, r *http.Request) {
	// MVP: devolvemos hasta 200 pendientes
	msgs, err := h.repo.ListPendingOutbox(r.Context(), 200)
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}
	writeJSON(w, http.StatusOK, msgs)
}

// @Summary Expire holds now
// @Description Ejecuta el job de expiración de holds de forma inmediata (operación interna).
// @Tags booking
// @Accept json
// @Produce json
// @Success 200 {object} ExpireHoldsNowResponse
// @Failure 500 {object} ErrorResponse
// @Router /internal/holds/expire-now [post]
func (h *Handlers) ExpireHoldsNow(w http.ResponseWriter, r *http.Request) {
	res, err := h.expireHoldsUC.Execute(r.Context(), usecases.ExpireHoldsInput{Limit: 500})
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}
	writeJSON(w, http.StatusOK, ExpireHoldsNowResponse{Expired: res.Expired})
}
