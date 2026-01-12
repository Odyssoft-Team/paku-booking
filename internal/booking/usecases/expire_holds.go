package usecases

import (
	"context"
	"time"

	"paku-booking/internal/booking"
)

type ExpireHoldsUseCase struct {
	Repo booking.Repository
	Now  func() time.Time
}

type ExpireHoldsInput struct {
	Limit int
}

type ExpireHoldsResult struct {
	Expired int
}

func (uc ExpireHoldsUseCase) Execute(ctx context.Context, in ExpireHoldsInput) (ExpireHoldsResult, error) {
	now := time.Now().UTC()
	if uc.Now != nil {
		now = uc.Now().UTC()
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 200
	}

	expiredCount := 0

	err := uc.Repo.Tx(ctx, func(ctx context.Context, tx booking.TxRepo) error {
		holds, err := tx.ListExpiredActiveHolds(ctx, now, limit)
		if err != nil {
			return err
		}

		for _, h := range holds {
			// liberar capacity
			if err := tx.ReleaseCapacity(ctx, h.ServiceID, h.LocationID, h.Date, h.Slot, h.Qty); err != nil {
				return err
			}

			// marcar hold expirado
			if err := tx.UpdateHoldStatus(ctx, h.ID, booking.HoldExpired, now); err != nil {
				return err
			}

			// outbox
			if err := insertOutbox(ctx, tx, now, booking.EventHoldExpired, "hold", h.ID, booking.HoldExpiredData{
				HoldID:     h.ID,
				ServiceID:  h.ServiceID,
				LocationID: h.LocationID,
				Date:       h.Date.Format(booking.DateLayout),
				Slot:       h.Slot,
				Qty:        h.Qty,
			}); err != nil {
				return err
			}

			expiredCount++
		}

		return nil
	})

	if err != nil {
		return ExpireHoldsResult{}, err
	}

	return ExpireHoldsResult{Expired: expiredCount}, nil
}
