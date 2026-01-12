package postgres

import (
	"context"
	"database/sql"
	"time"

	"paku-booking/internal/booking"
)

func (r *Repo) ListPendingOutbox(ctx context.Context, limit int) ([]booking.OutboxMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, type, payload, status, attempts, last_error, created_at, sent_at
		FROM outbox
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, string(booking.OutboxPending), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []booking.OutboxMessage
	for rows.Next() {
		var msg booking.OutboxMessage
		var typeStr, statusStr string
		var sentAt sql.NullTime

		err := rows.Scan(
			&msg.ID,
			&typeStr,
			&msg.Payload,
			&statusStr,
			&msg.Attempts,
			&msg.LastError,
			&msg.CreatedAt,
			&sentAt,
		)
		if err != nil {
			return nil, err
		}

		msg.Type = booking.EventType(typeStr)
		msg.Status = booking.OutboxStatus(statusStr)
		if sentAt.Valid {
			msg.SentAt = &sentAt.Time
		}

		result = append(result, msg)
	}

	return result, rows.Err()
}

func (r *Repo) MarkOutboxSent(ctx context.Context, outboxID string, sentAt time.Time) error {
	query := `
		UPDATE outbox
		SET status = $1, sent_at = $2, last_error = ''
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, string(booking.OutboxSent), sentAt, outboxID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return booking.ErrNotFound
	}

	return nil
}

func (r *Repo) MarkOutboxFailed(ctx context.Context, outboxID string, attemptAt time.Time, errMsg string) error {
	query := `
		UPDATE outbox
		SET status = $1, attempts = attempts + 1, last_error = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, string(booking.OutboxFailed), errMsg, outboxID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return booking.ErrNotFound
	}

	return nil
}

func (tx *txRepo) InsertOutbox(ctx context.Context, msg booking.OutboxMessage) error {
	query := `
		INSERT INTO outbox (id, type, payload, status, attempts, last_error, created_at, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := tx.tx.ExecContext(ctx, query,
		msg.ID,
		string(msg.Type),
		msg.Payload,
		string(msg.Status),
		msg.Attempts,
		msg.LastError,
		msg.CreatedAt,
		msg.SentAt,
	)
	return err
}
