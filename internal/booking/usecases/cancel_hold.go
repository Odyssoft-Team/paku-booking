package usecases

import (
	"context"
	"time"

	"paku-booking/internal/booking"
)

type CancelHoldUseCase struct {
	Repo booking.Repository
	Now  func() time.Time
}

type CancelHoldInput struct {
	HoldID string
}

func (uc CancelHoldUseCase) Execute(ctx context.Context, in CancelHoldInput) error {
	if in.HoldID == "" {
		return booking.ErrInvalidInput
	}

	now := time.Now().UTC()
	if uc.Now != nil {
		now = uc.Now().UTC()
	}

	// MVP note:
	// Ideal: GetHoldForUpdate dentro de tx (cuando sea Postgres).
	hold, err := uc.Repo.GetHold(ctx, in.HoldID)
	if err != nil {
		return err
	}
	if hold == nil {
		return booking.ErrNotFound
	}

	// Idempotencia:
	// - Si no está ACTIVE, no hacemos nada.
	if hold.Status != booking.HoldActive {
		return nil
	}

	// Decidimos si esto será "cancel" o "expire"
	expired := now.After(hold.ExpiresAt)

	return uc.Repo.Tx(ctx, func(ctx context.Context, tx booking.TxRepo) error {
		// 1) liberar capacity (siempre, porque estaba ACTIVE)
		if err := tx.ReleaseCapacity(ctx, hold.ServiceID, hold.LocationID, hold.Date, hold.Slot, hold.Qty); err != nil {
			return err
		}

		// 2) marcar estado final
		newStatus := booking.HoldCanceled
		eventType := booking.EventHoldCanceled

		if expired {
			newStatus = booking.HoldExpired
			eventType = booking.EventHoldExpired
		}

		if err := tx.UpdateHoldStatus(ctx, hold.ID, newStatus, now); err != nil {
			return err
		}

		// 3) outbox según caso
		if expired {
			return insertOutbox(ctx, tx, now, eventType, "hold", hold.ID, booking.HoldExpiredData{
				HoldID:     hold.ID,
				ServiceID:  hold.ServiceID,
				LocationID: hold.LocationID,
				Date:       hold.Date.Format(booking.DateLayout),
				Slot:       hold.Slot,
				Qty:        hold.Qty,
			})
		}

		return insertOutbox(ctx, tx, now, eventType, "hold", hold.ID, booking.HoldCanceledData{
			HoldID:     hold.ID,
			ServiceID:  hold.ServiceID,
			LocationID: hold.LocationID,
			Date:       hold.Date.Format(booking.DateLayout),
			Slot:       hold.Slot,
			Qty:        hold.Qty,
		})
	})
}
