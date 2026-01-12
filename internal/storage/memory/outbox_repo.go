package memory

import (
	"context"
	"time"

	"paku-booking/internal/booking"
)

func (r *Repo) ListPendingOutbox(ctx context.Context, limit int) ([]booking.OutboxMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if limit <= 0 {
		limit = 50
	}

	out := make([]booking.OutboxMessage, 0, limit)
	for _, id := range r.outboxOrder {
		m := r.outbox[id]
		if m == nil {
			continue
		}
		if m.Status == booking.OutboxPending {
			out = append(out, *m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *Repo) MarkOutboxSent(ctx context.Context, outboxID string, sentAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m := r.outbox[outboxID]
	if m == nil {
		return booking.ErrNotFound
	}
	m.Status = booking.OutboxSent
	m.SentAt = &sentAt
	m.LastError = ""
	return nil
}

func (r *Repo) MarkOutboxFailed(ctx context.Context, outboxID string, attemptAt time.Time, errMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m := r.outbox[outboxID]
	if m == nil {
		return booking.ErrNotFound
	}
	m.Status = booking.OutboxFailed
	m.Attempts++
	m.LastError = errMsg
	return nil
}

func (tx *txRepo) InsertOutbox(ctx context.Context, msg booking.OutboxMessage) error {
	r := (*Repo)(tx)

	cp := msg
	r.outbox[msg.ID] = &cp
	r.outboxOrder = append(r.outboxOrder, msg.ID)
	return nil
}
