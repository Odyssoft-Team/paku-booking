package usecases

import (
	"context"
	"time"

	"paku-booking/internal/booking"
)

type AvailabilityUseCase struct {
	Repo booking.Repository
}

type AvailabilityInput struct {
	ServiceID  string
	LocationID string
	From       time.Time
	To         time.Time
}

func (uc AvailabilityUseCase) Execute(ctx context.Context, in AvailabilityInput) ([]booking.DaySlot, error) {
	if in.ServiceID == "" {
		return nil, booking.ErrInvalidInput
	}

	from := normalizeDate(in.From)
	to := normalizeDate(in.To)
	if to.Before(from) {
		return nil, booking.ErrInvalidInput
	}

	q := booking.AvailabilityQuery{
		ServiceID:  in.ServiceID,
		LocationID: in.LocationID,
		From:       from,
		To:         to,
	}
	return uc.Repo.ListAvailability(ctx, q)
}
