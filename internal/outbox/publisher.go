package outbox

import (
	"context"

	"paku-booking/internal/booking"
)

// Publisher envía un mensaje de outbox hacia "algo":
// - HTTP webhook
// - Kafka/Rabbit (futuro)
// - Log (dev)
type Publisher interface {
	Publish(ctx context.Context, msg booking.OutboxMessage) error
}

// Dev publisher: solo loguea, no envía a ningún lado.
type NopPublisher struct{}

func (NopPublisher) Publish(ctx context.Context, msg booking.OutboxMessage) error {
	return nil
}
