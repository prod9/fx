package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
)

// PublishRaw sends a pre-marshaled payload on a channel by name — the untyped floor
// beneath Publish. It validates the name and size limit, resolves the driver (WithDriver
// → PUBSUB_URL → Postgres over the data context), and hands the send to it. Whether the
// publish rides an ambient transaction is driver-specific; prefer publishing
// out-of-band, after the business commit — the one ordering correct on every driver.
func PublishRaw(ctx context.Context, name string, payload string) error {
	if err := validateName(name); err != nil {
		return err
	}
	if len(payload) >= maxPayloadBytes {
		return fmt.Errorf("%w: %d bytes", ErrPayloadTooLarge, len(payload))
	}

	d, err := resolveDriver(ctx)
	if err != nil {
		return err
	}
	return d.Publish(ctx, name, payload)
}

// Publish marshals payload as JSON of T and sends it on ch. The JSON boundary is the one
// sanctioned any-shaped seam; treat T like any wire schema across deploys (add fields,
// never repurpose them).
func Publish[T any](ctx context.Context, ch Channel[T], payload T) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return PublishRaw(ctx, ch.name, string(data))
}
