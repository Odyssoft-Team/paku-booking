package booking

import "time"

// DaySlot es el inventario de cupos por día + slot (AM/PM)
type DaySlot struct {
	ServiceID  string
	LocationID string // opcional ("" si no aplica)
	Date       time.Time
	Slot       Slot
	Total      int
	Reserved   int
	UpdatedAt  time.Time
}

func (ds DaySlot) Available() int {
	return ds.Total - ds.Reserved
}

// Hold: bloqueo temporal previo al pago (consume cupo mientras dura)
type Hold struct {
	ID         string
	ServiceID  string
	LocationID string
	Date       time.Time
	Slot       Slot
	Qty        int

	Status    HoldStatus
	ExpiresAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Booking confirmado (post pago)
type Booking struct {
	ID        string
	HoldID    string
	PaymentID string

	ServiceID  string
	LocationID string
	Date       time.Time
	Slot       Slot
	Qty        int

	Status    BookingStatus
	CreatedAt time.Time
}

// --- Events/Outbox (MVP en memoria) ---

/* type EventType string

const (
	EventHoldCreated      EventType = "HOLD_CREATED"
	EventHoldCanceled     EventType = "HOLD_CANCELED"
	EventHoldExpired      EventType = "HOLD_EXPIRED"
	EventBookingConfirmed EventType = "BOOKING_CONFIRMED"
) */

type OutboxStatus string

const (
	OutboxPending OutboxStatus = "PENDING"
	OutboxSent    OutboxStatus = "SENT"
	OutboxFailed  OutboxStatus = "FAILED"
)

type OutboxMessage struct {
	ID        string
	Type      EventType
	Payload   []byte // JSON
	Status    OutboxStatus
	CreatedAt time.Time
	SentAt    *time.Time
	Attempts  int
	LastError string
}
