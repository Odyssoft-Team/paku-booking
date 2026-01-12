package outbox

import (
	"net/http"
	"os"
	"strings"
	"time"

	"paku-booking/internal/shared"
)

type PublisherConfig struct {
	Env string

	// Prod / HTTP
	HTTPEndpoint string
	HTTPAPIKey   string
	HTTPTimeout  time.Duration

	// Dev
	DevFail bool
}

func PublisherFromConfig(cfg PublisherConfig, logger shared.Logger) Publisher {
	env := strings.ToLower(strings.TrimSpace(cfg.Env))

	// Prod: HTTP publisher
	if env == "prod" || env == "production" {
		endpoint := cfg.HTTPEndpoint
		if endpoint == "" {
			endpoint = os.Getenv("OUTBOX_HTTP_ENDPOINT")
		}
		apiKey := cfg.HTTPAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("OUTBOX_HTTP_API_KEY")
		}

		timeout := cfg.HTTPTimeout
		if timeout <= 0 {
			timeout = 8 * time.Second
		}

		return &HTTPPublisher{
			Endpoint: endpoint,
			APIKey:   apiKey,
			Client:   &http.Client{Timeout: timeout},
			Logger:   logger,
		}
	}

	// Dev: logger publisher
	fail := cfg.DevFail
	if !fail {
		v := strings.TrimSpace(os.Getenv("OUTBOX_DEV_FAIL"))
		fail = (v == "1" || strings.EqualFold(v, "true"))
	}

	return DevPublisher{
		Logger: logger,
		Fail:   fail,
	}
}
