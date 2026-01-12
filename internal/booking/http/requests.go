package httpapi

import "paku-booking/internal/booking"

// Archivo: internal/booking/http/requests.go
//
// Solo structs de request/response JSON.
// (Nota: en handlers.go ya los tenías. Si los pasas aquí, handlers.go queda más limpio.)

type AvailabilityResponse struct {
	ServiceID  string            `json:"service_id"`
	LocationID string            `json:"location_id,omitempty"`
	From       string            `json:"from"` // YYYY-MM-DD
	To         string            `json:"to"`   // YYYY-MM-DD
	Days       []AvailabilityDay `json:"days"`
}

type AvailabilityDay struct {
	Date  string                `json:"date"` // YYYY-MM-DD
	Slots []AvailabilityDaySlot `json:"slots"`
}

type AvailabilityDaySlot struct {
	Slot      booking.Slot `json:"slot"` // AM/PM
	Total     int          `json:"total"`
	Reserved  int          `json:"reserved"`
	Available int          `json:"available"`
}

type CreateHoldRequest struct {
	ServiceID  string `json:"service_id"`
	LocationID string `json:"location_id,omitempty"`
	Date       string `json:"date"` // YYYY-MM-DD
	Slot       string `json:"slot"` // AM|PM
	Qty        int    `json:"qty"`  // default 1
}

type CreateHoldResponse struct {
	HoldID    string `json:"hold_id"`
	ExpiresAt string `json:"expires_at"` // RFC3339
}

type ConfirmBookingRequest struct {
	HoldID    string `json:"hold_id"`
	PaymentID string `json:"payment_id"`
}

type ConfirmBookingResponse struct {
	BookingID string `json:"booking_id"`
	Status    string `json:"status"`
}

// Admin

type AdminSetCapacityRequest struct {
	ServiceID  string `json:"service_id"`
	LocationID string `json:"location_id,omitempty"`
	Date       string `json:"date"` // YYYY-MM-DD
	Slot       string `json:"slot"` // AM|PM
	Total      int    `json:"total"`
}

type AdminAdjustCapacityRequest struct {
	ServiceID  string `json:"service_id"`
	LocationID string `json:"location_id,omitempty"`
	From       string `json:"from"` // YYYY-MM-DD
	To         string `json:"to"`   // YYYY-MM-DD
	Slot       string `json:"slot"` // AM|PM (si vacío: ambos)
	Delta      int    `json:"delta"`
}

type AdminCloseDaysRequest struct {
	ServiceID  string `json:"service_id"`
	LocationID string `json:"location_id,omitempty"`
	From       string `json:"from"` // YYYY-MM-DD
	To         string `json:"to"`   // YYYY-MM-DD
	Slot       string `json:"slot"` // AM|PM (si vacío: ambos)
}

// Internal/MVP helpers

type ExpireHoldsNowResponse struct {
	Expired int `json:"expired"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
