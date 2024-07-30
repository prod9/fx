package main

import (
	"testing"

	"fx.prodigy9.co/config"
	"github.com/stretchr/testify/require"
)

// Implemented here to not complicate the config package with a bunch of fs abstractions,
// for now.
func TestMain_Vars(t *testing.T) {
	src := config.Configure()
	config.Set(src, Name1, "Forced")

	require.Equal(t, "Forced", config.Get(src, Name1))
	require.Equal(t, "EnvLocal", config.Get(src, Name2))
	require.Equal(t, "Env", config.Get(src, Name3))
	require.Equal(t, "EnvLocalParent", config.Get(src, Name4))
	require.Equal(t, "EnvParent", config.Get(src, Name5))
	require.Equal(t, "Default6", config.Get(src, Name6))
}
