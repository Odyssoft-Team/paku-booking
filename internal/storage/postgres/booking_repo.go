package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"paku-booking/internal/booking"
)

// ---------- booking.Repository (lecturas sin Tx) ----------

func (r *Repo) ListAvailability(ctx context.Context, q booking.AvailabilityQuery) ([]booking.DaySlot, error) {
	from := booking.NormalizeDate(q.From)
	to := booking.NormalizeDate(q.To)

	query := `
		SELECT service_id, location_id, date, slot, total, reserved, updated_at
		FROM day_slots
		WHERE service_id = $1 
			AND location_id = $2
			AND date >= $3
			AND date <= $4
		ORDER BY date, slot
	`

	rows, err := r.db.QueryContext(ctx, query, q.ServiceID, q.LocationID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []booking.DaySlot
	for rows.Next() {
		var ds booking.DaySlot
		var slotStr string
		err := rows.Scan(
			&ds.ServiceID,
			&ds.LocationID,
			&ds.Date,
			&slotStr,
			&ds.Total,
			&ds.Reserved,
			&ds.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ds.Slot = booking.Slot(slotStr)
		result = append(result, ds)
	}

	return result, rows.Err()
}

func (r *Repo) GetHold(ctx context.Context, holdID string) (*booking.Hold, error) {
	query := `
		SELECT id, service_id, location_id, date, slot, qty, status, expires_at, created_at, updated_at
		FROM holds
		WHERE id = $1
	`

	var h booking.Hold
	var slotStr, statusStr string

	err := r.db.QueryRowContext(ctx, query, holdID).Scan(
		&h.ID,
		&h.ServiceID,
		&h.LocationID,
		&h.Date,
		&slotStr,
		&h.Qty,
		&statusStr,
		&h.ExpiresAt,
		&h.CreatedAt,
		&h.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	h.Slot = booking.Slot(slotStr)
	h.Status = booking.HoldStatus(statusStr)
	return &h, nil
}

// ---------- booking.TxRepo (operaciones dentro de Tx) ----------

func (tx *txRepo) UpsertDaySlot(ctx context.Context, slot booking.DaySlot) error {
	date := booking.NormalizeDate(slot.Date)

	// Verificar si existe
	var existingReserved int
	checkQuery := `SELECT reserved FROM day_slots WHERE service_id=$1 AND location_id=$2 AND date=$3 AND slot=$4`
	err := tx.tx.QueryRowContext(ctx, checkQuery, slot.ServiceID, slot.LocationID, date, string(slot.Slot)).Scan(&existingReserved)

	if errors.Is(err, sql.ErrNoRows) {
		// Insert
		insertQuery := `
			INSERT INTO day_slots (service_id, location_id, date, slot, total, reserved, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = tx.tx.ExecContext(ctx, insertQuery,
			slot.ServiceID, slot.LocationID, date, string(slot.Slot),
			slot.Total, slot.Reserved, slot.UpdatedAt,
		)
		return err
	}
	if err != nil {
		return err
	}

	// Update - validar total >= reserved
	if slot.Total < existingReserved {
		return booking.ErrBelowReserved
	}

	updateQuery := `
		UPDATE day_slots 
		SET total = $1, updated_at = $2
		WHERE service_id = $3 AND location_id = $4 AND date = $5 AND slot = $6
	`
	_, err = tx.tx.ExecContext(ctx, updateQuery,
		slot.Total, slot.UpdatedAt,
		slot.ServiceID, slot.LocationID, date, string(slot.Slot),
	)
	return err
}

func (tx *txRepo) GetDaySlot(ctx context.Context, serviceID, locationID string, date time.Time, slotType booking.Slot) (*booking.DaySlot, error) {
	date = booking.NormalizeDate(date)

	query := `
		SELECT service_id, location_id, date, slot, total, reserved, updated_at
		FROM day_slots
		WHERE service_id = $1 AND location_id = $2 AND date = $3 AND slot = $4
	`

	var ds booking.DaySlot
	var slotStr string

	err := tx.tx.QueryRowContext(ctx, query, serviceID, locationID, date, string(slotType)).Scan(
		&ds.ServiceID,
		&ds.LocationID,
		&ds.Date,
		&slotStr,
		&ds.Total,
		&ds.Reserved,
		&ds.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	ds.Slot = booking.Slot(slotStr)
	return &ds, nil
}

func (tx *txRepo) ReserveCapacity(ctx context.Context, serviceID, locationID string, date time.Time, slotType booking.Slot, qty int) error {
	if qty <= 0 {
		qty = 1
	}

	date = booking.NormalizeDate(date)

	// Anti-overbooking: UPDATE condicional que verifica capacidad disponible
	query := `
		UPDATE day_slots
		SET reserved = reserved + $1, updated_at = NOW()
		WHERE service_id = $2 
			AND location_id = $3 
			AND date = $4 
			AND slot = $5
			AND reserved + $1 <= total
	`

	result, err := tx.tx.ExecContext(ctx, query, qty, serviceID, locationID, date, string(slotType))
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		// Verificar si el slot existe
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM day_slots WHERE service_id=$1 AND location_id=$2 AND date=$3 AND slot=$4)`
		if err := tx.tx.QueryRowContext(ctx, checkQuery, serviceID, locationID, date, string(slotType)).Scan(&exists); err != nil {
			return err
		}

		if !exists {
			return booking.ErrDaySlotMissing
		}
		return booking.ErrNoCapacity
	}

	return nil
}

func (tx *txRepo) ReleaseCapacity(ctx context.Context, serviceID, locationID string, date time.Time, slotType booking.Slot, qty int) error {
	if qty <= 0 {
		qty = 1
	}

	date = booking.NormalizeDate(date)

	query := `
		UPDATE day_slots
		SET reserved = GREATEST(0, reserved - $1), updated_at = NOW()
		WHERE service_id = $2 AND location_id = $3 AND date = $4 AND slot = $5
	`

	_, err := tx.tx.ExecContext(ctx, query, qty, serviceID, locationID, date, string(slotType))
	return err
}

func (tx *txRepo) SetCapacity(ctx context.Context, serviceID, locationID string, date time.Time, slotType booking.Slot, total int) error {
	if total < 0 {
		total = 0
	}

	date = booking.NormalizeDate(date)

	// Intentar insert primero
	insertQuery := `
		INSERT INTO day_slots (service_id, location_id, date, slot, total, reserved, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, NOW())
		ON CONFLICT (service_id, location_id, date, slot) DO NOTHING
	`
	_, err := tx.tx.ExecContext(ctx, insertQuery, serviceID, locationID, date, string(slotType), total)
	if err != nil {
		return err
	}

	// Update solo si total >= reserved
	updateQuery := `
		UPDATE day_slots
		SET total = $1, updated_at = NOW()
		WHERE service_id = $2 AND location_id = $3 AND date = $4 AND slot = $5
			AND $1 >= reserved
	`
	result, err := tx.tx.ExecContext(ctx, updateQuery, total, serviceID, locationID, date, string(slotType))
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		// Verificar si falló por la constraint
		var reserved int
		checkQuery := `SELECT reserved FROM day_slots WHERE service_id=$1 AND location_id=$2 AND date=$3 AND slot=$4`
		if err := tx.tx.QueryRowContext(ctx, checkQuery, serviceID, locationID, date, string(slotType)).Scan(&reserved); err == nil {
			if total < reserved {
				return booking.ErrBelowReserved
			}
		}
	}

	return nil
}

func (tx *txRepo) AdjustCapacityRange(ctx context.Context, serviceID, locationID string, from, to time.Time, slots []booking.Slot, delta int) error {
	from = booking.NormalizeDate(from)
	to = booking.NormalizeDate(to)

	if len(slots) == 0 {
		slots = []booking.Slot{booking.SlotAM, booking.SlotPM}
	}

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		for _, sl := range slots {
			// Insertar si no existe
			insertQuery := `
				INSERT INTO day_slots (service_id, location_id, date, slot, total, reserved, updated_at)
				VALUES ($1, $2, $3, $4, 0, 0, NOW())
				ON CONFLICT (service_id, location_id, date, slot) DO NOTHING
			`
			_, err := tx.tx.ExecContext(ctx, insertQuery, serviceID, locationID, d, string(sl))
			if err != nil {
				return err
			}

			// Update con validación
			updateQuery := `
				UPDATE day_slots
				SET total = GREATEST(0, total + $1), updated_at = NOW()
				WHERE service_id = $2 AND location_id = $3 AND date = $4 AND slot = $5
					AND GREATEST(0, total + $1) >= reserved
			`
			result, err := tx.tx.ExecContext(ctx, updateQuery, delta, serviceID, locationID, d, string(sl))
			if err != nil {
				return err
			}

			affected, _ := result.RowsAffected()
			if affected == 0 {
				return booking.ErrBelowReserved
			}
		}
	}

	return nil
}

func (tx *txRepo) CloseDaysRange(ctx context.Context, serviceID, locationID string, from, to time.Time, slots []booking.Slot) error {
	from = booking.NormalizeDate(from)
	to = booking.NormalizeDate(to)

	if len(slots) == 0 {
		slots = []booking.Slot{booking.SlotAM, booking.SlotPM}
	}

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		for _, sl := range slots {
			// Insertar si no existe (con total=0)
			insertQuery := `
				INSERT INTO day_slots (service_id, location_id, date, slot, total, reserved, updated_at)
				VALUES ($1, $2, $3, $4, 0, 0, NOW())
				ON CONFLICT (service_id, location_id, date, slot) DO NOTHING
			`
			_, err := tx.tx.ExecContext(ctx, insertQuery, serviceID, locationID, d, string(sl))
			if err != nil {
				return err
			}

			// Update a 0 solo si reserved = 0
			updateQuery := `
				UPDATE day_slots
				SET total = 0, updated_at = NOW()
				WHERE service_id = $1 AND location_id = $2 AND date = $3 AND slot = $4
					AND reserved = 0
			`
			result, err := tx.tx.ExecContext(ctx, updateQuery, serviceID, locationID, d, string(sl))
			if err != nil {
				return err
			}

			affected, _ := result.RowsAffected()
			if affected == 0 {
				// Verificar si hay reservas
				var reserved int
				checkQuery := `SELECT reserved FROM day_slots WHERE service_id=$1 AND location_id=$2 AND date=$3 AND slot=$4`
				if err := tx.tx.QueryRowContext(ctx, checkQuery, serviceID, locationID, d, string(sl)).Scan(&reserved); err == nil {
					if reserved > 0 {
						return booking.ErrBelowReserved
					}
				}
			}
		}
	}

	return nil
}

// --- Holds ---

func (tx *txRepo) InsertHold(ctx context.Context, h booking.Hold) error {
	query := `
		INSERT INTO holds (id, service_id, location_id, date, slot, qty, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := tx.tx.ExecContext(ctx, query,
		h.ID, h.ServiceID, h.LocationID, booking.NormalizeDate(h.Date),
		string(h.Slot), h.Qty, string(h.Status), h.ExpiresAt, h.CreatedAt, h.UpdatedAt,
	)
	return err
}

func (tx *txRepo) UpdateHoldStatus(ctx context.Context, holdID string, status booking.HoldStatus, updatedAt time.Time) error {
	query := `UPDATE holds SET status = $1, updated_at = $2 WHERE id = $3`

	result, err := tx.tx.ExecContext(ctx, query, string(status), updatedAt, holdID)
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

func (tx *txRepo) ListExpiredActiveHolds(ctx context.Context, now time.Time, limit int) ([]booking.Hold, error) {
	if limit <= 0 {
		limit = 200
	}

	query := `
		SELECT id, service_id, location_id, date, slot, qty, status, expires_at, created_at, updated_at
		FROM holds
		WHERE status = $1 AND expires_at < $2
		ORDER BY expires_at
		LIMIT $3
	`

	rows, err := tx.tx.QueryContext(ctx, query, string(booking.HoldActive), now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []booking.Hold
	for rows.Next() {
		var h booking.Hold
		var slotStr, statusStr string

		err := rows.Scan(
			&h.ID, &h.ServiceID, &h.LocationID, &h.Date,
			&slotStr, &h.Qty, &statusStr, &h.ExpiresAt,
			&h.CreatedAt, &h.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		h.Slot = booking.Slot(slotStr)
		h.Status = booking.HoldStatus(statusStr)
		result = append(result, h)
	}

	return result, rows.Err()
}

// --- Bookings ---

func (tx *txRepo) InsertBooking(ctx context.Context, b booking.Booking) error {
	query := `
		INSERT INTO bookings (id, hold_id, payment_id, service_id, location_id, date, slot, qty, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := tx.tx.ExecContext(ctx, query,
		b.ID, b.HoldID, b.PaymentID, b.ServiceID, b.LocationID,
		booking.NormalizeDate(b.Date), string(b.Slot), b.Qty,
		string(b.Status), b.CreatedAt,
	)
	return err
}
