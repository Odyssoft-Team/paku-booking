package jobs

import (
	"context"
	"time"

	"paku-booking/internal/booking"
	"paku-booking/internal/booking/usecases"
	"paku-booking/internal/shared"
)

// Job: expira holds vencidos.
// Importante: este job NO publica eventos; solo ejecuta el usecase ExpireHolds,
// que ya escribe outbox (EventHoldExpired). El dispatcher se encarga de enviarlos.
type HoldExpirer struct {
	Repo booking.Repository

	Interval time.Duration
	Limit    int

	Clock  shared.Clock
	Logger shared.Logger
}

func (j *HoldExpirer) Run(ctx context.Context) {
	if j.Repo == nil {
		if j.Logger != nil {
			j.Logger.Errorf("hold expirer: missing Repo")
		}
		return
	}

	interval := j.Interval
	if interval <= 0 {
		interval = 1 * time.Minute
	}

	limit := j.Limit
	if limit <= 0 {
		limit = 200
	}

	clk := j.Clock
	if clk == nil {
		clk = shared.RealClock{}
	}

	t := time.NewTicker(interval)
	defer t.Stop()

	if j.Logger != nil {
		j.Logger.Infof("hold expirer started (interval=%s limit=%d)", interval, limit)
	}

	uc := usecases.ExpireHoldsUseCase{
		Repo: j.Repo,
		Now:  clk.Now,
	}

	for {
		select {
		case <-ctx.Done():
			if j.Logger != nil {
				j.Logger.Infof("hold expirer stopped")
			}
			return
		case <-t.C:
			res, err := uc.Execute(ctx, usecases.ExpireHoldsInput{Limit: limit})
			if err != nil {
				if j.Logger != nil {
					j.Logger.Errorf("hold expirer tick error: %v", err)
				}
				continue
			}
			if res.Expired > 0 && j.Logger != nil {
				j.Logger.Infof("hold expirer: expired=%d", res.Expired)
			}
		}
	}
}
