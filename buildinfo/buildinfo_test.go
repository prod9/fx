package buildinfo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromContext_ReturnsTheContextsInfo(t *testing.T) {
	want := Info{Name: "api", Version: "v1.2.3", Revision: "abc123"}
	ctx := NewContext(context.Background(), want)

	require.Equal(t, want, FromContext(ctx))
}

func TestFromContext_FallsBackToGet(t *testing.T) {
	got := FromContext(context.Background())

	require.Equal(t, Get(), got)
	require.NotEmpty(t, got.Name)
}
