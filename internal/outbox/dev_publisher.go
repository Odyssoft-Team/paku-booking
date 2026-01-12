package outbox

import (
	"context"

	"paku-booking/internal/booking"
	"paku-booking/internal/shared"
)

type DevPublisher struct {
	Logger shared.Logger
	Fail   bool
}

func (p DevPublisher) Publish(ctx context.Context, msg booking.OutboxMessage) error {
	_ = ctx

	if p.Logger != nil {
		// Log “defensivo”: no asumimos campos que no existan
		p.Logger.Infof(
			"DEV EVENT outbox_id=%s type=%s payload=%s",
			msg.ID, msg.Type, string(msg.Payload),
		)
	}

	if p.Fail {
		return ErrDevPublishForced
	}
	return nil
}
