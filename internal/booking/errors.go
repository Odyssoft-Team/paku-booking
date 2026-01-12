package booking

import "errors"

// Errores de dominio (negocio). Úsalos en repos y/o usecases.
// En HTTP se mapearán a status codes.

var (
	// Validación / input
	ErrInvalidInput = errors.New("invalid input")

	// Recursos
	ErrNotFound = errors.New("not found")

	// Capacidad
	ErrNoCapacity     = errors.New("no capacity")
	ErrDaySlotMissing = errors.New("day slot missing") // capacidad no configurada
	ErrBelowReserved  = errors.New("below reserved")   // no permitir total < reserved

	// Estado / lifecycle
	ErrInvalidState = errors.New("invalid state")
	ErrExpired      = errors.New("expired")
)
