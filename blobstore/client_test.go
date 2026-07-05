package blobstore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecureFromScheme(t *testing.T) {
	cases := []struct {
		scheme string
		secure bool
		ok     bool
	}{
		{"s3", true, true},    // live default, unchanged
		{"https", true, true}, // explicit TLS endpoint
		{"http", false, true}, // local blobserver
		{"ftp", false, false}, // unknown scheme denied
		{"", false, false},    // missing scheme denied
	}

	for _, c := range cases {
		secure, err := secureFromScheme(c.scheme)
		if !c.ok {
			require.Error(t, err, c.scheme)
			continue
		}
		require.NoError(t, err, c.scheme)
		require.Equal(t, c.secure, secure, c.scheme)
	}
}
