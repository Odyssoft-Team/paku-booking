package outbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"paku-booking/internal/booking"
	"paku-booking/internal/shared"
)

type HTTPPublisher struct {
	Endpoint string // ej: https://paku-core/api/events (o un webhook interno)
	APIKey   string // opcional
	Client   *http.Client
	Logger   shared.Logger
}

func (p *HTTPPublisher) Publish(ctx context.Context, msg booking.OutboxMessage) error {
	if p.Endpoint == "" {
		return fmt.Errorf("http publisher: empty endpoint")
	}

	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(msg.Payload))
	if err != nil {
		return fmt.Errorf("http publisher: new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Idempotency-Key", msg.ID)
	if p.APIKey != "" {
		req.Header.Set("X-API-Key", p.APIKey)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http publisher: do: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if p.Logger != nil {
			p.Logger.Debugf("outbox http publish ok: id=%s status=%d", msg.ID, res.StatusCode)
		}
		return nil
	}

	// Lee un poco del body para diagnóstico (sin reventar memoria)
	body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
	return fmt.Errorf("http publisher: status=%d body=%q", res.StatusCode, string(body))
}
