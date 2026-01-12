package booking

import (
	"encoding/json"
	"time"
)

// Archivo: internal/booking/events.go
//
// Objetivo: definir tipos de eventos + helper para crear payload JSON.
// En MVP, el outbox guarda Payload como []byte (JSON). Más adelante, esto
// te permite mantener el contrato estable aunque cambies el publisher.
//
// Nota: la estructura de eventos está pensada para consumers externos.
// Manténlo simple y consistente: ids + datos clave del hold/booking.

type EventType string

const (
	EventHoldCreated      EventType = "HOLD_CREATED"
	EventHoldCanceled     EventType = "HOLD_CANCELED"
	EventHoldExpired      EventType = "HOLD_EXPIRED"
	EventBookingConfirmed EventType = "BOOKING_CONFIRMED"
)

// Envelopa estándar para outbox (útil si luego publicas a broker/webhook).
type EventEnvelope struct {
	ID         string    `json:"id"` // id del mensaje/evento (no del hold)
	Type       EventType `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`

	// Referencias útiles para tracing/consumo
	AggregateType string `json:"aggregate_type"` // "hold" | "booking"
	AggregateID   string `json:"aggregate_id"`   // hold_id | booking_id

	// Payload específico del evento
	Data any `json:"data"`
}

// ---- Payloads por evento ----

type HoldCreatedData struct {
	HoldID     string `json:"hold_id"`
	ServiceID  string `json:"service_id"`
	LocationID string `json:"location_id,omitempty"`
	Date       string `json:"date"` // YYYY-MM-DD
	Slot       Slot   `json:"slot"` // AM/PM
	Qty        int    `json:"qty"`
	ExpiresAt  string `json:"expires_at"` // RFC3339
}

type HoldCanceledData struct {
	HoldID     string `json:"hold_id"`
	ServiceID  string `json:"service_id"`
	LocationID string `json:"location_id,omitempty"`
	Date       string `json:"date"` // YYYY-MM-DD
	Slot       Slot   `json:"slot"` // AM/PM
	Qty        int    `json:"qty"`
}

type HoldExpiredData struct {
	HoldID     string `json:"hold_id"`
	ServiceID  string `json:"service_id"`
	LocationID string `json:"location_id,omitempty"`
	Date       string `json:"date"` // YYYY-MM-DD
	Slot       Slot   `json:"slot"` // AM/PM
	Qty        int    `json:"qty"`
}

type BookingConfirmedData struct {
	BookingID  string `json:"booking_id"`
	HoldID     string `json:"hold_id"`
	PaymentID  string `json:"payment_id"`
	ServiceID  string `json:"service_id"`
	LocationID string `json:"location_id,omitempty"`
	Date       string `json:"date"` // YYYY-MM-DD
	Slot       Slot   `json:"slot"` // AM/PM
	Qty        int    `json:"qty"`
}

// BuildEnvelope crea el envelope + lo serializa a JSON para guardarlo en outbox.
func BuildEnvelope(
	eventID string,
	typ EventType,
	occurredAt time.Time,
	aggregateType string,
	aggregateID string,
	data any,
) ([]byte, error) {
	env := EventEnvelope{
		ID:            eventID,
		Type:          typ,
		OccurredAt:    occurredAt.UTC(),
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Data:          data,
	}
	return json.Marshal(env)
}
