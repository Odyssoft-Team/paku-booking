package httpapi

import (
	"net/http"
	"time"

	"paku-booking/internal/booking"
	"paku-booking/internal/booking/usecases"
)

func (h *Handlers) GetAvailability(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("service_id")
	locationID := r.URL.Query().Get("location_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if serviceID == "" || fromStr == "" || toStr == "" {
		writeErr(w, http.StatusBadRequest, "missing required query params: service_id, from, to")
		return
	}

	from, err := time.Parse(booking.DateLayout, fromStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid from date (expected YYYY-MM-DD)")
		return
	}
	to, err := time.Parse(booking.DateLayout, toStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid to date (expected YYYY-MM-DD)")
		return
	}

	// el usecase normaliza también, pero validamos acá por UX.
	from = booking.NormalizeDate(from)
	to = booking.NormalizeDate(to)
	if to.Before(from) {
		writeErr(w, http.StatusBadRequest, "to must be >= from")
		return
	}

	slots, err := h.availabilityUC.Execute(r.Context(), usecases.AvailabilityInput{
		ServiceID:  serviceID,
		LocationID: locationID,
		From:       from,
		To:         to,
	})
	if err != nil {
		he := mapError(err)
		writeErr(w, he.status, he.msg)
		return
	}

	// Convertimos []booking.DaySlot -> AvailabilityResponse (por día)
	resp := AvailabilityResponse{
		ServiceID:  serviceID,
		LocationID: locationID,
		From:       from.Format(booking.DateLayout),
		To:         to.Format(booking.DateLayout),
		Days:       groupSlotsByDay(slots),
	}

	writeJSON(w, http.StatusOK, resp)
}

func groupSlotsByDay(slots []booking.DaySlot) []AvailabilityDay {
	byDate := map[string][]AvailabilityDaySlot{}

	for _, s := range slots {
		ds := s.Date.Format(booking.DateLayout)
		byDate[ds] = append(byDate[ds], AvailabilityDaySlot{
			Slot:      s.Slot,
			Total:     s.Total,
			Reserved:  s.Reserved,
			Available: s.Available(),
		})
	}

	// Mantener orden por fecha (simple)
	out := make([]AvailabilityDay, 0, len(byDate))
	for d := range byDate {
		out = append(out, AvailabilityDay{
			Date:  d,
			Slots: byDate[d],
		})
	}

	// Orden por fecha asc (YYYY-MM-DD ordena lexicográfico)
	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Date < out[i].Date {
				out[i], out[j] = out[j], out[i]
			}
		}
	}

	return out
}
