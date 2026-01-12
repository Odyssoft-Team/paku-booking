package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open abre una conexión a PostgreSQL usando pgx/v5 a través de database/sql.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	// Configurar pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return db, nil
}

// SQL para crear tablas (usar en migraciones)
const Schema = `
-- day_slots: inventario de cupos por servicio/ubicación/fecha/slot
CREATE TABLE IF NOT EXISTS day_slots (
	service_id TEXT NOT NULL,
	location_id TEXT NOT NULL DEFAULT '',
	date DATE NOT NULL,
	slot TEXT NOT NULL, -- 'AM' | 'PM'
	total INT NOT NULL DEFAULT 0,
	reserved INT NOT NULL DEFAULT 0,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (service_id, location_id, date, slot),
	CHECK (reserved >= 0),
	CHECK (total >= reserved)
);

CREATE INDEX IF NOT EXISTS idx_day_slots_service_date 
	ON day_slots(service_id, location_id, date);

-- holds: bloqueos temporales pre-pago
CREATE TABLE IF NOT EXISTS holds (
	id TEXT PRIMARY KEY,
	service_id TEXT NOT NULL,
	location_id TEXT NOT NULL DEFAULT '',
	date DATE NOT NULL,
	slot TEXT NOT NULL,
	qty INT NOT NULL,
	status TEXT NOT NULL, -- 'ACTIVE' | 'CANCELED' | 'EXPIRED' | 'CONFIRMED'
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_holds_status_expires 
	ON holds(status, expires_at) WHERE status = 'ACTIVE';

-- bookings: reservas confirmadas post-pago
CREATE TABLE IF NOT EXISTS bookings (
	id TEXT PRIMARY KEY,
	hold_id TEXT NOT NULL,
	payment_id TEXT NOT NULL,
	service_id TEXT NOT NULL,
	location_id TEXT NOT NULL DEFAULT '',
	date DATE NOT NULL,
	slot TEXT NOT NULL,
	qty INT NOT NULL,
	status TEXT NOT NULL, -- 'CONFIRMED' | 'CANCELED'
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bookings_hold ON bookings(hold_id);

-- outbox: eventos para publicación asíncrona
CREATE TABLE IF NOT EXISTS outbox (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	payload JSONB NOT NULL,
	status TEXT NOT NULL, -- 'PENDING' | 'SENT' | 'FAILED'
	attempts INT NOT NULL DEFAULT 0,
	last_error TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	sent_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_outbox_status_created 
	ON outbox(status, created_at) WHERE status = 'PENDING';
`
