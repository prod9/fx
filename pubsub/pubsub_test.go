package pubsub

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateName(t *testing.T) {
	cases := []struct {
		name  string
		input string
		valid bool
	}{
		{"simple", "orders_changed", true},
		{"leading underscore", "_private", true},
		{"digits after letter", "ch1", true},
		{"exactly 63 bytes", strings.Repeat("a", 63), true},
		{"empty", "", false},
		{"64 bytes", strings.Repeat("a", 64), false},
		{"leading digit", "1channel", false},
		{"hyphen", "orders-changed", false},
		{"colon", "orders:42", false},
		{"space", "orders changed", false},
		{"dot", "orders.changed", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateName(c.input)
			if c.valid {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrInvalidName)
			}
		})
	}
}

func TestNewChannelPanicsOnInvalidName(t *testing.T) {
	require.Panics(t, func() { NewChannel[struct{}]("bad name") })
	require.NotPanics(t, func() { NewChannel[struct{}]("good_name") })
}

func TestPublishRawRejectsOversizePayload(t *testing.T) {
	// The size guard must reject before any DB access, so a background context with no
	// data connection still surfaces the error rather than panicking on a missing DB.
	err := PublishRaw(context.Background(), "valid_name", strings.Repeat("x", maxPayloadBytes))
	require.ErrorIs(t, err, ErrPayloadTooLarge)
}

func TestPublishRawRejectsInvalidName(t *testing.T) {
	err := PublishRaw(context.Background(), "bad name", "x")
	require.ErrorIs(t, err, ErrInvalidName)
}

func TestSubscribeRawRequiresDatabase(t *testing.T) {
	_, _, err := SubscribeRaw(context.Background(), "valid_name")
	require.ErrorIs(t, err, ErrNoDatabase)
}

// guard against accidental removal of the exclusive-limit boundary.
func TestPayloadLimitIsExclusive(t *testing.T) {
	require.True(t, errors.Is(PublishRaw(context.Background(), "n", strings.Repeat("x", maxPayloadBytes)), ErrPayloadTooLarge),
		"a payload of exactly maxPayloadBytes must be rejected")
}
