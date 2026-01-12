package booking

import (
	"fmt"
	"strings"
	"time"
)

type Slot string

const (
	SlotAM Slot = "AM"
	SlotPM Slot = "PM"
)

func ParseSlot(s string) (Slot, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "AM":
		return SlotAM, nil
	case "PM":
		return SlotPM, nil
	default:
		return "", fmt.Errorf("invalid slot: %q (expected AM|PM)", s)
	}
}

func (s Slot) IsValid() bool { return s == SlotAM || s == SlotPM }

type HoldStatus string

const (
	HoldActive    HoldStatus = "ACTIVE"
	HoldCanceled  HoldStatus = "CANCELED"
	HoldExpired   HoldStatus = "EXPIRED"
	HoldConfirmed HoldStatus = "CONFIRMED"
)

type BookingStatus string

const (
	BookingConfirmed BookingStatus = "CONFIRMED"
)

// Fecha normalizada: siempre usar date-only (00:00) en UTC para claves.
// Si prefieres local, cámbialo, pero sé consistente.
func NormalizeDate(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

// YYYY-MM-DD
const DateLayout = "2006-01-02"
