package httpapi

// @tag.name booking
// @tag.description Operaciones del dominio "booking": disponibilidad, holds, bookings, administración y operativas internas.

import (
	"encoding/json"
	"net/http"

	"paku-booking/internal/booking"
	"paku-booking/internal/booking/usecases"
	"paku-booking/internal/shared"
)

type HandlersDeps struct {
	Env    string
	Repo   booking.Repository
	Clock  shared.Clock
	Logger shared.Logger

	AvailabilityUC     usecases.AvailabilityUseCase
	CreateHoldUC       usecases.CreateHoldUseCase
	CancelHoldUC       usecases.CancelHoldUseCase
	ConfirmBookingUC   usecases.ConfirmBookingUseCase
	ExpireHoldsUC      usecases.ExpireHoldsUseCase
	AdminSetCapacityUC usecases.AdminSetCapacityUseCase
	AdminAdjustCapUC   usecases.AdminAdjustCapacityUseCase
	AdminCloseDaysUC   usecases.AdminCloseDaysUseCase
}

type Handlers struct {
	env    string
	repo   booking.Repository
	clock  shared.Clock
	logger shared.Logger

	availabilityUC     usecases.AvailabilityUseCase
	createHoldUC       usecases.CreateHoldUseCase
	cancelHoldUC       usecases.CancelHoldUseCase
	confirmBookingUC   usecases.ConfirmBookingUseCase
	expireHoldsUC      usecases.ExpireHoldsUseCase
	adminSetCapacityUC usecases.AdminSetCapacityUseCase
	adminAdjustCapUC   usecases.AdminAdjustCapacityUseCase
	adminCloseDaysUC   usecases.AdminCloseDaysUseCase
}

func NewHandlers(d HandlersDeps) *Handlers {
	return &Handlers{
		env:    d.Env,
		repo:   d.Repo,
		clock:  d.Clock,
		logger: d.Logger,

		availabilityUC:     d.AvailabilityUC,
		createHoldUC:       d.CreateHoldUC,
		cancelHoldUC:       d.CancelHoldUC,
		confirmBookingUC:   d.ConfirmBookingUC,
		expireHoldsUC:      d.ExpireHoldsUC,
		adminSetCapacityUC: d.AdminSetCapacityUC,
		adminAdjustCapUC:   d.AdminAdjustCapUC,
		adminCloseDaysUC:   d.AdminCloseDaysUC,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}
