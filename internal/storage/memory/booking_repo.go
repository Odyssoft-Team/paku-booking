package memory

import (
	"context"
	"time"

	"paku-booking/internal/booking"
)

type slotKey struct {
	serviceID  string
	locationID string
	date       string // YYYY-MM-DD
	slot       booking.Slot
}

func makeKey(serviceID, locationID string, date time.Time, slot booking.Slot) slotKey {
	return slotKey{
		serviceID:  serviceID,
		locationID: locationID,
		date:       date.Format(booking.DateLayout),
		slot:       slot,
	}
}

// ---------- booking.Repository (lecturas) ----------

func (r *Repo) ListAvailability(ctx context.Context, q booking.AvailabilityQuery) ([]booking.DaySlot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	from := booking.NormalizeDate(q.From)
	to := booking.NormalizeDate(q.To)

	out := make([]booking.DaySlot, 0)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		for _, sl := range []booking.Slot{booking.SlotAM, booking.SlotPM} {
			k := makeKey(q.ServiceID, q.LocationID, d, sl)
			ds, ok := r.daySlots[k]
			if !ok {
				continue
			}
			out = append(out, *ds)
		}
	}
	return out, nil
}

func (r *Repo) GetHold(ctx context.Context, holdID string) (*booking.Hold, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	h, ok := r.holds[holdID]
	if !ok {
		return nil, nil
	}
	cp := *h
	return &cp, nil
}

// ---------- booking.TxRepo (writes dentro de Tx) ----------

func (tx *txRepo) UpsertDaySlot(ctx context.Context, slot booking.DaySlot) error {
	r := (*Repo)(tx)

	k := makeKey(slot.ServiceID, slot.LocationID, slot.Date, slot.Slot)
	existing, ok := r.daySlots[k]
	if !ok {
		cp := slot
		r.daySlots[k] = &cp
		return nil
	}

	if slot.Total < existing.Reserved {
		return booking.ErrBelowReserved
	}

	existing.Total = slot.Total
	existing.UpdatedAt = slot.UpdatedAt
	return nil
}

func (tx *txRepo) GetDaySlot(ctx context.Context, serviceID, locationID string, date time.Time, slotType booking.Slot) (*booking.DaySlot, error) {
	r := (*Repo)(tx)

	k := makeKey(serviceID, locationID, date, slotType)
	ds, ok := r.daySlots[k]
	if !ok {
		return nil, nil
	}
	cp := *ds
	return &cp, nil
}

func (tx *txRepo) ReserveCapacity(ctx context.Context, serviceID, locationID string, date time.Time, slotType booking.Slot, qty int) error {
	r := (*Repo)(tx)
	if qty <= 0 {
		qty = 1
	}

	k := makeKey(serviceID, locationID, date, slotType)
	ds, ok := r.daySlots[k]
	if !ok {
		return booking.ErrDaySlotMissing
	}
	if ds.Reserved+qty > ds.Total {
		return booking.ErrNoCapacity
	}

	ds.Reserved += qty
	ds.UpdatedAt = time.Now().UTC()
	return nil
}

func (tx *txRepo) ReleaseCapacity(ctx context.Context, serviceID, locationID string, date time.Time, slotType booking.Slot, qty int) error {
	r := (*Repo)(tx)
	if qty <= 0 {
		qty = 1
	}

	k := makeKey(serviceID, locationID, date, slotType)
	ds, ok := r.daySlots[k]
	if !ok {
		return nil
	}

	ds.Reserved -= qty
	if ds.Reserved < 0 {
		ds.Reserved = 0
	}
	ds.UpdatedAt = time.Now().UTC()
	return nil
}

func (tx *txRepo) SetCapacity(ctx context.Context, serviceID, locationID string, date time.Time, slotType booking.Slot, total int) error {
	r := (*Repo)(tx)

	if total < 0 {
		total = 0
	}

	k := makeKey(serviceID, locationID, date, slotType)
	ds, ok := r.daySlots[k]
	if !ok {
		r.daySlots[k] = &booking.DaySlot{
			ServiceID:  serviceID,
			LocationID: locationID,
			Date:       booking.NormalizeDate(date),
			Slot:       slotType,
			Total:      total,
			Reserved:   0,
			UpdatedAt:  time.Now().UTC(),
		}
		return nil
	}

	if total < ds.Reserved {
		return booking.ErrBelowReserved
	}

	ds.Total = total
	ds.UpdatedAt = time.Now().UTC()
	return nil
}

func (tx *txRepo) AdjustCapacityRange(ctx context.Context, serviceID, locationID string, from, to time.Time, slots []booking.Slot, delta int) error {
	r := (*Repo)(tx)

	from = booking.NormalizeDate(from)
	to = booking.NormalizeDate(to)
	if len(slots) == 0 {
		slots = []booking.Slot{booking.SlotAM, booking.SlotPM}
	}

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		for _, sl := range slots {
			k := makeKey(serviceID, locationID, d, sl)
			ds, ok := r.daySlots[k]
			if !ok {
				ds = &booking.DaySlot{
					ServiceID:  serviceID,
					LocationID: locationID,
					Date:       d,
					Slot:       sl,
					Total:      0,
					Reserved:   0,
					UpdatedAt:  time.Now().UTC(),
				}
				r.daySlots[k] = ds
			}

			newTotal := ds.Total + delta
			if newTotal < 0 {
				newTotal = 0
			}
			if newTotal < ds.Reserved {
				return booking.ErrBelowReserved
			}
			ds.Total = newTotal
			ds.UpdatedAt = time.Now().UTC()
		}
	}

	return nil
}

func (tx *txRepo) CloseDaysRange(ctx context.Context, serviceID, locationID string, from, to time.Time, slots []booking.Slot) error {
	r := (*Repo)(tx)

	from = booking.NormalizeDate(from)
	to = booking.NormalizeDate(to)
	if len(slots) == 0 {
		slots = []booking.Slot{booking.SlotAM, booking.SlotPM}
	}

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		for _, sl := range slots {
			k := makeKey(serviceID, locationID, d, sl)
			ds, ok := r.daySlots[k]
			if !ok {
				r.daySlots[k] = &booking.DaySlot{
					ServiceID:  serviceID,
					LocationID: locationID,
					Date:       d,
					Slot:       sl,
					Total:      0,
					Reserved:   0,
					UpdatedAt:  time.Now().UTC(),
				}
				continue
			}
			if ds.Reserved > 0 {
				return booking.ErrBelowReserved
			}
			ds.Total = 0
			ds.UpdatedAt = time.Now().UTC()
		}
	}

	return nil
}

func (tx *txRepo) InsertHold(ctx context.Context, h booking.Hold) error {
	r := (*Repo)(tx)
	cp := h
	r.holds[h.ID] = &cp
	return nil
}

func (tx *txRepo) UpdateHoldStatus(ctx context.Context, holdID string, status booking.HoldStatus, updatedAt time.Time) error {
	r := (*Repo)(tx)
	h, ok := r.holds[holdID]
	if !ok {
		return booking.ErrNotFound
	}
	h.Status = status
	h.UpdatedAt = updatedAt
	return nil
}

func (tx *txRepo) ListExpiredActiveHolds(ctx context.Context, now time.Time, limit int) ([]booking.Hold, error) {
	r := (*Repo)(tx)

	if limit <= 0 {
		limit = 200
	}

	out := make([]booking.Hold, 0, limit)
	for _, h := range r.holds {
		if h.Status == booking.HoldActive && now.After(h.ExpiresAt) {
			out = append(out, *h)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (tx *txRepo) InsertBooking(ctx context.Context, b booking.Booking) error {
	r := (*Repo)(tx)
	cp := b
	r.bookings[b.ID] = &cp
	return nil
}
