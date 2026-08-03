package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"fx.prodigy9.co/data"
)

// PublishRaw sends a pre-marshaled payload on a channel by name — the untyped floor
// beneath Publish. It runs through data.Exec, so it rides whatever transaction context it
// is called in: inside data.Run it joins the parent tx and fires on that commit; on its
// own it runs a BEGIN/pg_notify/COMMIT of its own. Prefer publishing out-of-band, after
// the business commit — coupling a NOTIFY into the business tx lets a full notify queue
// roll it back.
func PublishRaw(ctx context.Context, name string, payload string) error {
	if err := validateName(name); err != nil {
		return err
	}
	if len(payload) >= maxPayloadBytes {
		return fmt.Errorf("%w: %d bytes", ErrPayloadTooLarge, len(payload))
	}

	return data.Exec(ctx, `SELECT pg_notify($1, $2)`, name, payload)
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
