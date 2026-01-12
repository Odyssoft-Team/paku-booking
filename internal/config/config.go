package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env string // dev|prod

	Port string

	HoldTTL time.Duration

	// Outbox / jobs (por ahora solo placeholders, pero config ya queda lista)
	OutboxDispatcherInterval time.Duration
	HoldExpirerInterval      time.Duration
}

func Load() Config {
	env := getenv("ENV", "dev")
	port := getenv("PORT", "8080")

	ttlMin := getenvInt("HOLD_TTL_MINUTES", 15)
	holdTTL := time.Duration(ttlMin) * time.Minute

	outboxSec := getenvInt("OUTBOX_DISPATCH_INTERVAL_SECONDS", 5)
	holdExpMin := getenvInt("HOLD_EXPIRE_INTERVAL_MINUTES", 1)

	return Config{
		Env:                      env,
		Port:                     port,
		HoldTTL:                  holdTTL,
		OutboxDispatcherInterval: time.Duration(outboxSec) * time.Second,
		HoldExpirerInterval:      time.Duration(holdExpMin) * time.Minute,
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
