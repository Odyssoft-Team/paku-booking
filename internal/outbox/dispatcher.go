package outbox

import (
	"context"
	"time"

	"paku-booking/internal/booking"
	"paku-booking/internal/shared"
)

type Dispatcher struct {
	Repo      booking.Repository
	Publisher Publisher

	Interval  time.Duration
	BatchSize int

	Clock  shared.Clock
	Logger shared.Logger
}

// Run corre hasta que el ctx se cancele.
func (d *Dispatcher) Run(ctx context.Context) {
	if d.Repo == nil || d.Publisher == nil {
		if d.Logger != nil {
			d.Logger.Errorf("outbox dispatcher: missing Repo or Publisher")
		}
		return
	}

	interval := d.Interval
	if interval <= 0 {
		interval = 5 * time.Second
	}

	batch := d.BatchSize
	if batch <= 0 {
		batch = 50
	}

	clk := d.Clock
	if clk == nil {
		clk = shared.RealClock{}
	}

	t := time.NewTicker(interval)
	defer t.Stop()

	if d.Logger != nil {
		d.Logger.Infof("outbox dispatcher started (interval=%s batch=%d)", interval, batch)
	}

	for {
		select {
		case <-ctx.Done():
			if d.Logger != nil {
				d.Logger.Infof("outbox dispatcher stopped")
			}
			return
		case <-t.C:
			d.tick(ctx, clk, batch)
		}
	}
}

func (d *Dispatcher) tick(ctx context.Context, clk shared.Clock, batch int) {
	msgs, err := d.Repo.ListPendingOutbox(ctx, batch)
	if err != nil {
		if d.Logger != nil {
			d.Logger.Errorf("outbox list pending error: %v", err)
		}
		return
	}
	if len(msgs) == 0 {
		return
	}

	now := clk.Now()

	for _, m := range msgs {
		// publish
		if err := d.Publisher.Publish(ctx, m); err != nil {
			_ = d.Repo.MarkOutboxFailed(ctx, m.ID, now, err.Error())
			if d.Logger != nil {
				d.Logger.Errorf("outbox publish failed id=%s err=%v", m.ID, err)
			}
			continue
		}

		// mark sent
		_ = d.Repo.MarkOutboxSent(ctx, m.ID, now)
		if d.Logger != nil {
			d.Logger.Debugf("outbox sent id=%s type=%s", m.ID, m.Type)
		}
	}
}
