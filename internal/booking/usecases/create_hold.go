package usecases

import (
	"context"
	"time"

	"paku-booking/internal/booking"
)

type CreateHoldUseCase struct {
	Repo    booking.Repository
	HoldTTL time.Duration
	Now     func() time.Time // opcional
}

type CreateHoldInput struct {
	ServiceID  string
	LocationID string
	Date       time.Time
	Slot       booking.Slot
	Qty        int
}

type CreateHoldResult struct {
	HoldID    string
	ExpiresAt time.Time
}

func (uc CreateHoldUseCase) Execute(ctx context.Context, in CreateHoldInput) (CreateHoldResult, error) {
	if in.ServiceID == "" || !in.Slot.IsValid() {
		return CreateHoldResult{}, booking.ErrInvalidInput
	}

	qty := clampQty(in.Qty)

	now := time.Now().UTC()
	if uc.Now != nil {
		now = uc.Now().UTC()
	}

	ttl := uc.HoldTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	date := normalizeDate(in.Date)

	// Micro-ajuste A: no permitir reservar días en el pasado.
	// (trabajamos por día, no por hora, así que "hoy" es válido).
	today := normalizeDate(now)
	if date.Before(today) {
		return CreateHoldResult{}, booking.ErrInvalidInput
	}

	expiresAt := now.Add(ttl)
	holdID := newID()

	err := uc.Repo.Tx(ctx, func(ctx context.Context, tx booking.TxRepo) error {
		// 1) consumir capacidad (anti-overbooking)
		if err := tx.ReserveCapacity(ctx, in.ServiceID, in.LocationID, date, in.Slot, qty); err != nil {
			return err
		}

		// 2) crear hold
		h := booking.Hold{
			ID:         holdID,
			ServiceID:  in.ServiceID,
			LocationID: in.LocationID,
			Date:       date,
			Slot:       in.Slot,
			Qty:        qty,
			Status:     booking.HoldActive,
			ExpiresAt:  expiresAt,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := tx.InsertHold(ctx, h); err != nil {
			return err
		}

		// 3) outbox
		return insertOutbox(ctx, tx, now, booking.EventHoldCreated, "hold", holdID, booking.HoldCreatedData{
			HoldID:     holdID,
			ServiceID:  in.ServiceID,
			LocationID: in.LocationID,
			Date:       date.Format(booking.DateLayout),
			Slot:       in.Slot,
			Qty:        qty,
			ExpiresAt:  expiresAt.Format(time.RFC3339),
		})
	})

	if err != nil {
		return CreateHoldResult{}, err
	}

	return CreateHoldResult{HoldID: holdID, ExpiresAt: expiresAt}, nil
}
