package httpapi

import (
	"net/http"
	"os"
	"time"

	"paku-booking/internal/booking"
	"paku-booking/internal/booking/usecases"
	"paku-booking/internal/shared"
	"paku-booking/internal/storage/memory"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type RouterOptions struct {
	// Si no lo pasas, se usa memory.NewRepo()
	Repo booking.Repository

	Clock  shared.Clock
	Logger shared.Logger

	HoldTTL time.Duration
	Env     string
}

// @Summary Health check
// @Description Estado de salud simple del servicio.
// @Tags booking
// @Accept json
// @Produce json
// @Success 200 {string} string "ok"
// @Router /health [get]
func Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func NewRouter(opts RouterOptions) http.Handler {
	r := chi.NewRouter()

	// Middleware base
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	// Health
	r.Get("/health", Health)

	// ---- deps (inyección) ----
	repo := opts.Repo
	if repo == nil {
		repo = memory.NewRepo()
	}

	clk := opts.Clock
	if clk == nil {
		clk = shared.RealClock{}
	}

	log := opts.Logger
	if log == nil {
		log = shared.NewStdLogger()
	}

	holdTTL := opts.HoldTTL
	if holdTTL <= 0 {
		holdTTL = 15 * time.Minute
	}

	env := opts.Env
	if env == "" {
		env = os.Getenv("ENV")
	}

	// ---- usecases ----
	availabilityUC := usecases.AvailabilityUseCase{Repo: repo}

	createHoldUC := usecases.CreateHoldUseCase{
		Repo:    repo,
		HoldTTL: holdTTL,
		Now:     clk.Now,
	}

	cancelHoldUC := usecases.CancelHoldUseCase{
		Repo: repo,
		Now:  clk.Now,
	}

	confirmBookingUC := usecases.ConfirmBookingUseCase{
		Repo: repo,
		Now:  clk.Now,
	}

	expireHoldsUC := usecases.ExpireHoldsUseCase{
		Repo: repo,
		Now:  clk.Now,
	}

	adminSetCapacityUC := usecases.AdminSetCapacityUseCase{
		Repo: repo,
		Now:  clk.Now,
	}

	adminAdjustCapUC := usecases.AdminAdjustCapacityUseCase{
		Repo: repo,
	}

	adminCloseDaysUC := usecases.AdminCloseDaysUseCase{
		Repo: repo,
	}

	// ---- handlers ----
	h := NewHandlers(HandlersDeps{
		Env:    env,
		Repo:   repo,
		Clock:  clk,
		Logger: log,

		AvailabilityUC:     availabilityUC,
		CreateHoldUC:       createHoldUC,
		CancelHoldUC:       cancelHoldUC,
		ConfirmBookingUC:   confirmBookingUC,
		ExpireHoldsUC:      expireHoldsUC,
		AdminSetCapacityUC: adminSetCapacityUC,
		AdminAdjustCapUC:   adminAdjustCapUC,
		AdminCloseDaysUC:   adminCloseDaysUC,
	})

	// ---- routes ----
	r.Get("/availability", h.GetAvailability)

	r.Post("/holds", h.CreateHold)
	r.Post("/holds/{id}/cancel", h.CancelHold)

	r.Post("/bookings/confirm", h.ConfirmBooking)

	r.Route("/admin", func(ar chi.Router) {
		ar.Put("/capacity", h.AdminSetCapacity)
		ar.Post("/capacity/adjust", h.AdminAdjustCapacity)
		ar.Post("/capacity/close-days", h.AdminCloseDays)
	})

	// Ops / internal
	r.Route("/internal", func(ir chi.Router) {
		ir.Post("/holds/expire-now", h.ExpireHoldsNow)
		ir.Get("/outbox/pending", h.GetOutboxPending)
	})

	return r
}
