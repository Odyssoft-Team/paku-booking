package usecases

import (
	"context"
	"time"

	"paku-booking/internal/booking"
)

type AdminSetCapacityUseCase struct {
	Repo booking.Repository
	Now  func() time.Time
}

type AdminSetCapacityInput struct {
	ServiceID  string
	LocationID string
	Date       time.Time
	Slot       booking.Slot
	Total      int
}

func (uc AdminSetCapacityUseCase) Execute(ctx context.Context, in AdminSetCapacityInput) error {
	if in.ServiceID == "" || !in.Slot.IsValid() || in.Total < 0 {
		return booking.ErrInvalidInput
	}

	date := normalizeDate(in.Date)

	_ = time.Now().UTC()
	if uc.Now != nil {
		_ = uc.Now().UTC()
	}

	return uc.Repo.Tx(ctx, func(ctx context.Context, tx booking.TxRepo) error {
		return tx.SetCapacity(ctx, in.ServiceID, in.LocationID, date, in.Slot, in.Total)
	})
}
