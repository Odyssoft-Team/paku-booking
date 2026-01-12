package memory

import (
	"context"
	"sync"

	"paku-booking/internal/booking"
)

type Repo struct {
	mu sync.Mutex

	daySlots map[slotKey]*booking.DaySlot
	holds    map[string]*booking.Hold
	bookings map[string]*booking.Booking

	outbox      map[string]*booking.OutboxMessage
	outboxOrder []string
}

func NewRepo() *Repo {
	return &Repo{
		daySlots:    make(map[slotKey]*booking.DaySlot),
		holds:       make(map[string]*booking.Hold),
		bookings:    make(map[string]*booking.Booking),
		outbox:      make(map[string]*booking.OutboxMessage),
		outboxOrder: make([]string, 0),
	}
}

// Tx implementa booking.Repository.Tx.
// En memoria: usamos mutex global para garantizar atomicidad.
func (r *Repo) Tx(ctx context.Context, fn func(ctx context.Context, tx booking.TxRepo) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return fn(ctx, (*txRepo)(r))
}

// txRepo es un "view" del Repo bajo el mismo lock.
// No agrega campos; solo implementa los métodos TxRepo.
type txRepo Repo

// (Esto asegura que Repo implementa booking.Repository)
var _ booking.Repository = (*Repo)(nil)

// (Esto asegura que txRepo implementa booking.TxRepo)
var _ booking.TxRepo = (*txRepo)(nil)

// Para evitar import "sync" no usado si lo editas luego.
var _ = sync.Mutex{}
