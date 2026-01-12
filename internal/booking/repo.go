package booking

import (
	"context"
	"time"
)

// Archivo: internal/booking/repo.go
//
// Punto medio: interfaces mínimas para que los usecases no dependan de Postgres.
// Nota: para el MVP puedes implementar TODO en un solo archivo postgres/booking_repo.go,
// pero esta interfaz te deja crecer sin dolor.
//
// Decisión: usamos Date normalizada (UTC 00:00) para claves y consultas.

type Repository interface {
	// Tx ejecuta fn dentro de una transacción.
	// Implementación Postgres: BEGIN -> fn -> COMMIT/ROLLBACK.
	Tx(ctx context.Context, fn func(ctx context.Context, tx TxRepo) error) error

	// Lecturas (no requieren Tx obligatoriamente)
	ListAvailability(ctx context.Context, q AvailabilityQuery) ([]DaySlot, error)
	GetHold(ctx context.Context, holdID string) (*Hold, error)

	// Outbox (normalmente no requiere Tx para lecturas)
	ListPendingOutbox(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkOutboxSent(ctx context.Context, outboxID string, sentAt time.Time) error
	MarkOutboxFailed(ctx context.Context, outboxID string, attemptAt time.Time, errMsg string) error
}

// TxRepo = subset de operaciones que deben ser atómicas.
// En general: crear hold + consumir cupo, confirmar booking, cancelar/expirar hold, ajustes admin.
type TxRepo interface {
	// --- Capacity / DaySlot ---
	UpsertDaySlot(ctx context.Context, slot DaySlot) error
	GetDaySlot(ctx context.Context, serviceID, locationID string, date time.Time, slotType Slot) (*DaySlot, error)

	// Consume capacity de forma atómica (anti-overbooking).
	// Debe fallar con ErrNoCapacity si reserved+qty > total.
	ReserveCapacity(ctx context.Context, serviceID, locationID string, date time.Time, slotType Slot, qty int) error

	// Libera capacity (para cancel/expire holds)
	ReleaseCapacity(ctx context.Context, serviceID, locationID string, date time.Time, slotType Slot, qty int) error

	// Ajustes admin
	SetCapacity(ctx context.Context, serviceID, locationID string, date time.Time, slotType Slot, total int) error
	AdjustCapacityRange(ctx context.Context, serviceID, locationID string, from, to time.Time, slots []Slot, delta int) error
	CloseDaysRange(ctx context.Context, serviceID, locationID string, from, to time.Time, slots []Slot) error

	// --- Holds ---
	InsertHold(ctx context.Context, h Hold) error
	UpdateHoldStatus(ctx context.Context, holdID string, status HoldStatus, updatedAt time.Time) error

	// Para job de expiración: obtiene holds activos expirados (limitado)
	ListExpiredActiveHolds(ctx context.Context, now time.Time, limit int) ([]Hold, error)

	// --- Bookings ---
	InsertBooking(ctx context.Context, b Booking) error

	// --- Outbox (en la misma Tx para atomicidad) ---
	InsertOutbox(ctx context.Context, msg OutboxMessage) error
}

// Query simple para availability
type AvailabilityQuery struct {
	ServiceID  string
	LocationID string // opcional
	From       time.Time
	To         time.Time
}

// OutboxMessage aquí usa el tipo del model (internal/booking/model.go).
// DaySlot, Hold, Booking igual.
