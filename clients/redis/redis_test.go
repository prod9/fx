package redis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientMemoizesPerURL(t *testing.T) {
	a, err := Client("redis://localhost:6379/0")
	require.NoError(t, err)
	b, err := Client("redis://localhost:6379/0")
	require.NoError(t, err)
	c, err := Client("redis://localhost:6379/1")
	require.NoError(t, err)

	require.Same(t, a, b)
	require.NotSame(t, a, c)
}

func TestClientRejectsBadURL(t *testing.T) {
	_, err := Client("not-a-redis-url")
	require.Error(t, err)
}
