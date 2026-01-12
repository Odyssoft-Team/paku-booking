package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"paku-booking/internal/booking"
	"paku-booking/internal/booking/usecases"
)

func (h *Handlers) AdminSetCapacity(w http.ResponseWriter, r *http.Request) {
	var req AdminSetCapacityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.ServiceID == "" || req.Date == "" || req.Slot == "" {
		writeErr(w, http.StatusBadRequest, "service_id, date, slot are required")
		return
	}
	if req.Total < 0 {
		writeErr(w, http.StatusBadRequest, "total must be >= 0")
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

	err = h.adminSetCapacityUC.Execute(r.Context(), usecases.AdminSetCapacityInput{
		ServiceID:  req.ServiceID,
		LocationID: req.LocationID,
		Date:       date,
		Slot:       slot,
		Total:      req.Total,
	})
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) AdminAdjustCapacity(w http.ResponseWriter, r *http.Request) {
	var req AdminAdjustCapacityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.ServiceID == "" || req.From == "" || req.To == "" {
		writeErr(w, http.StatusBadRequest, "service_id, from, to are required")
		return
	}

	from, err := time.Parse(booking.DateLayout, req.From)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid from (expected YYYY-MM-DD)")
		return
	}
	to, err := time.Parse(booking.DateLayout, req.To)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid to (expected YYYY-MM-DD)")
		return
	}
	from = booking.NormalizeDate(from)
	to = booking.NormalizeDate(to)
	if to.Before(from) {
		writeErr(w, http.StatusBadRequest, "to must be >= from")
		return
	}

	var slots []booking.Slot
	if req.Slot == "" {
		slots = []booking.Slot{booking.SlotAM, booking.SlotPM}
	} else {
		s, err := booking.ParseSlot(req.Slot)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		slots = []booking.Slot{s}
	}

	err = h.adminAdjustCapUC.Execute(r.Context(), usecases.AdminAdjustCapacityInput{
		ServiceID:  req.ServiceID,
		LocationID: req.LocationID,
		From:       from,
		To:         to,
		Slots:      slots,
		Delta:      req.Delta,
	})
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) AdminCloseDays(w http.ResponseWriter, r *http.Request) {
	var req AdminCloseDaysRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.ServiceID == "" || req.From == "" || req.To == "" {
		writeErr(w, http.StatusBadRequest, "service_id, from, to are required")
		return
	}

	from, err := time.Parse(booking.DateLayout, req.From)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid from (expected YYYY-MM-DD)")
		return
	}
	to, err := time.Parse(booking.DateLayout, req.To)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid to (expected YYYY-MM-DD)")
		return
	}
	from = booking.NormalizeDate(from)
	to = booking.NormalizeDate(to)
	if to.Before(from) {
		writeErr(w, http.StatusBadRequest, "to must be >= from")
		return
	}

	var slots []booking.Slot
	if req.Slot == "" {
		slots = []booking.Slot{booking.SlotAM, booking.SlotPM}
	} else {
		s, err := booking.ParseSlot(req.Slot)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		slots = []booking.Slot{s}
	}

	err = h.adminCloseDaysUC.Execute(r.Context(), usecases.AdminCloseDaysInput{
		ServiceID:  req.ServiceID,
		LocationID: req.LocationID,
		From:       from,
		To:         to,
		Slots:      slots,
	})
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
