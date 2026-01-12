package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"paku-booking/internal/booking"
)

type ConfirmBookingUseCase struct {
	Repo booking.Repository
	Now  func() time.Time
}

type ConfirmBookingInput struct {
	HoldID    string
	PaymentID string
}

type ConfirmBookingResult struct {
	BookingID string
}

func (uc ConfirmBookingUseCase) Execute(ctx context.Context, in ConfirmBookingInput) (ConfirmBookingResult, error) {
	if in.HoldID == "" || in.PaymentID == "" {
		return ConfirmBookingResult{}, booking.ErrInvalidInput
	}

	now := time.Now().UTC()
	if uc.Now != nil {
		now = uc.Now().UTC()
	}

	hold, err := uc.Repo.GetHold(ctx, in.HoldID)
	if err != nil {
		return ConfirmBookingResult{}, err
	}
	if hold == nil {
		return ConfirmBookingResult{}, booking.ErrNotFound
	}

	// Micro-ajuste B: manejos claros (MVP)
	// - Solo ACTIVE puede confirmarse
	// - Si ya expiró, ErrExpired
	// - Si está en otro estado, ErrInvalidState
	if hold.Status != booking.HoldActive {
		return ConfirmBookingResult{}, booking.ErrInvalidState
	}
	if now.After(hold.ExpiresAt) {
		return ConfirmBookingResult{}, booking.ErrExpired
	}

	// Idempotencia práctica:
	// bookingID determinístico para que reintentos no creen "múltiples bookings".
	bookingID := deterministicBookingID(hold.ID, in.PaymentID)

	err = uc.Repo.Tx(ctx, func(ctx context.Context, tx booking.TxRepo) error {
		// 1) marcar hold confirmado
		if err := tx.UpdateHoldStatus(ctx, hold.ID, booking.HoldConfirmed, now); err != nil {
			return err
		}

		// 2) insertar booking
		b := booking.Booking{
			ID:         bookingID,
			HoldID:     hold.ID,
			PaymentID:  in.PaymentID,
			ServiceID:  hold.ServiceID,
			LocationID: hold.LocationID,
			Date:       hold.Date,
			Slot:       hold.Slot,
			Qty:        hold.Qty,
			Status:     booking.BookingConfirmed,
			CreatedAt:  now,
		}
		if err := tx.InsertBooking(ctx, b); err != nil {
			return err
		}

		// 3) outbox
		return insertOutbox(ctx, tx, now, booking.EventBookingConfirmed, "booking", bookingID, booking.BookingConfirmedData{
			BookingID:  bookingID,
			HoldID:     hold.ID,
			PaymentID:  in.PaymentID,
			ServiceID:  hold.ServiceID,
			LocationID: hold.LocationID,
			Date:       hold.Date.Format(booking.DateLayout),
			Slot:       hold.Slot,
			Qty:        hold.Qty,
		})
	})

	if err != nil {
		return ConfirmBookingResult{}, err
	}

	return ConfirmBookingResult{BookingID: bookingID}, nil
}

// deterministicBookingID genera un ID estable para evitar duplicados por reintentos.
// No requiere cambios de repo ni tabla unique extra (aunque en Postgres igual conviene).
func deterministicBookingID(holdID, paymentID string) string {
	sum := sha256.Sum256([]byte("booking:" + holdID + ":" + paymentID))
	// 32 hex chars = 128 bits (suficiente para ID corto y estable)
	return hex.EncodeToString(sum[:16])
}
