package usecases

import (
	"context"
	"time"

	"paku-booking/internal/booking"
)

type AdminCloseDaysUseCase struct {
	Repo booking.Repository
}

type AdminCloseDaysInput struct {
	ServiceID  string
	LocationID string
	From       time.Time
	To         time.Time
	Slots      []booking.Slot
}

func (uc AdminCloseDaysUseCase) Execute(ctx context.Context, in AdminCloseDaysInput) error {
	if in.ServiceID == "" {
		return booking.ErrInvalidInput
	}

	from := normalizeDate(in.From)
	to := normalizeDate(in.To)
	if to.Before(from) {
		return booking.ErrInvalidInput
	}

	slots := in.Slots
	if len(slots) == 0 {
		slots = []booking.Slot{booking.SlotAM, booking.SlotPM}
	}
	for _, s := range slots {
		if !s.IsValid() {
			return booking.ErrInvalidInput
		}
	}

	return uc.Repo.Tx(ctx, func(ctx context.Context, tx booking.TxRepo) error {
		return tx.CloseDaysRange(ctx, in.ServiceID, in.LocationID, from, to, slots)
	})
}
