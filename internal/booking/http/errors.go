package httpapi

import (
	"errors"
	"net/http"

	"paku-booking/internal/booking"
)

type httpError struct {
	status int
	msg    string
}

func mapError(err error) httpError {
	// default
	out := httpError{
		status: http.StatusInternalServerError,
		msg:    "internal error",
	}
	if err == nil {
		return out
	}

	// Validación / input
	if errors.Is(err, booking.ErrInvalidInput) {
		return httpError{status: http.StatusBadRequest, msg: "invalid input"}
	}

	// Not found
	if errors.Is(err, booking.ErrNotFound) {
		return httpError{status: http.StatusNotFound, msg: "not found"}
	}

	// Capacidad
	if errors.Is(err, booking.ErrDaySlotMissing) {
		return httpError{status: http.StatusNotFound, msg: "capacity not configured for that day/slot"}
	}
	if errors.Is(err, booking.ErrNoCapacity) {
		return httpError{status: http.StatusConflict, msg: "no capacity available"}
	}
	if errors.Is(err, booking.ErrBelowReserved) {
		return httpError{status: http.StatusConflict, msg: "cannot reduce capacity below reserved"}
	}

	// Lifecycle/estado
	if errors.Is(err, booking.ErrExpired) {
		return httpError{status: http.StatusConflict, msg: "hold expired"}
	}
	if errors.Is(err, booking.ErrInvalidState) {
		return httpError{status: http.StatusConflict, msg: "invalid state"}
	}

	return out
}
