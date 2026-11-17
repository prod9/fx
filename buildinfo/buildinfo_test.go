package buildinfo

import (
	"context"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

// Resolution order: docs/spec/buildinfo.md.
func TestResolve(t *testing.T) {
	stamped := &debug.BuildInfo{
		Main:     debug.Module{Version: "v1.2.3"},
		Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "abc123"}},
	}
	devel := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}

	cases := []struct {
		name   string
		linked Info
		build  *debug.BuildInfo
		want   Info
	}{
		{"link-time values win", Info{Name: "api", Version: "v9.9.9"}, stamped,
			Info{Name: "api", Version: "v9.9.9", Revision: "abc123"}},
		{"go build info next", Info{}, stamped,
			Info{Name: "server", Version: "v1.2.3", Revision: "abc123"}},
		{"devel when unstamped", Info{}, devel,
			Info{Name: "server", Version: "(devel)"}},
		{"devel without build info", Info{}, nil,
			Info{Name: "server", Version: "(devel)"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, resolve(c.linked, c.build, "/usr/local/bin/server"))
		})
	}
}

func TestFromContext(t *testing.T) {
	require.Equal(t, Get(), FromContext(context.Background()))

	info := Get().WithName("worker-a")
	ctx := NewContext(context.Background(), info)
	require.Equal(t, "worker-a", FromContext(ctx).Name)
	require.Equal(t, Get().Version, FromContext(ctx).Version)
}
