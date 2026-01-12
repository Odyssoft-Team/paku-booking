package httpapi

import (
	"net/http"

	"paku-booking/internal/booking/usecases"
)

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

func (h *Handlers) ExpireHoldsNow(w http.ResponseWriter, r *http.Request) {
	res, err := h.expireHoldsUC.Execute(r.Context(), usecases.ExpireHoldsInput{Limit: 500})
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}
	writeJSON(w, http.StatusOK, ExpireHoldsNowResponse{Expired: res.Expired})
}
