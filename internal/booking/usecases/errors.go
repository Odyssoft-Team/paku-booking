package usecases

import (
	"context"
	"time"

	"paku-booking/internal/booking"

	"github.com/google/uuid"
)

func newID() string { return uuid.NewString() }

func normalizeDate(d time.Time) time.Time { return booking.NormalizeDate(d) }

func clampQty(q int) int {
	if q <= 0 {
		return 1
	}
	return q
}

func insertOutbox(
	ctx context.Context,
	tx booking.TxRepo,
	now time.Time,
	typ booking.EventType,
	aggregateType, aggregateID string,
	data any,
) error {
	eventID := newID()
	payload, err := booking.BuildEnvelope(eventID, typ, now, aggregateType, aggregateID, data)
	if err != nil {
		return err
	}

	msg := booking.OutboxMessage{
		ID:        eventID,
		Type:      typ,
		Payload:   payload,
		Status:    booking.OutboxPending,
		CreatedAt: now,
	}
	return tx.InsertOutbox(ctx, msg)
}
