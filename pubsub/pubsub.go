// Package pubsub is a typed pub/sub over pluggable drivers — Postgres LISTEN/NOTIFY by
// default, selected via PUBSUB_URL or an injected Driver. It is a latency optimization
// over polling, not a delivery guarantee: every driver is at-most-once and drops when no
// one listens or across a reconnect gap. Apps stay correct by treating a notification as
// "something changed, look now" and re-deriving from the table; anything needing
// guaranteed delivery uses worker, not pubsub. See docs/spec/pubsub.md.
package pubsub

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	// maxPayloadBytes is the Postgres NOTIFY payload ceiling; the limit is exclusive.
	maxPayloadBytes = 8000
	// maxNameBytes is Postgres NAMEDATALEN: identifiers past it truncate and silently
	// collide, so names are rejected rather than allowed to truncate.
	maxNameBytes = 63
)

var (
	ErrPayloadTooLarge = errors.New("pubsub: payload exceeds 8000 bytes")
	ErrInvalidName     = errors.New("pubsub: channel name must be a plain identifier of at most 63 bytes")

	namePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

// validateName is the package's one trust boundary: LISTEN cannot parameterize its
// channel name, so the name is interpolated into SQL. Whitelist plain identifiers and
// reject everything else, including names long enough to truncate-collide at NAMEDATALEN.
func validateName(name string) error {
	if len(name) == 0 || len(name) > maxNameBytes || !namePattern.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	return nil
}
