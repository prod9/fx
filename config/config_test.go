package config_test

import (
	"testing"

	"fx.prodigy9.co/config"
	"github.com/stretchr/testify/require"
)

// Alias lookup rules: docs/spec/configuration.md, "Aliases".
func TestGet_Aliases(t *testing.T) {
	cases := []struct {
		name   string
		values map[string]string
		want   string
		wantOK bool
	}{
		{"neither set", map[string]string{}, "default", false},
		{"canonical only", map[string]string{"NEW_NAME": "new"}, "new", true},
		{"alias only", map[string]string{"OLD_NAME": "old"}, "old", true},
		{"canonical wins over alias", map[string]string{"NEW_NAME": "new", "OLD_NAME": "old"}, "new", true},
		{"blank canonical falls through", map[string]string{"NEW_NAME": "  ", "OLD_NAME": "old"}, "old", true},
		{"first alias wins", map[string]string{"OLD_NAME": "old", "OLDER_NAME": "older"}, "old", true},
		{"later alias used", map[string]string{"OLDER_NAME": "older"}, "older", true},
	}

	v := config.StrDef("NEW_NAME", "default").WithAlias("OLD_NAME", "OLDER_NAME")
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := newSource(t, c.values)

			got, ok := config.GetOK(src, v)
			require.Equal(t, c.want, got)
			require.Equal(t, c.wantOK, ok)
			require.Equal(t, c.want, config.Get(src, v))
			require.Equal(t, c.want, config.GetAny(src, v))
		})
	}
}

func TestGet_AliasParsesWithVarType(t *testing.T) {
	v := config.IntDef("NEW_COUNT", 7).WithAlias("OLD_COUNT")

	require.Equal(t, 42, config.Get(newSource(t, map[string]string{"OLD_COUNT": "42"}), v))
	require.Equal(t, 7, config.Get(newSource(t, map[string]string{"OLD_COUNT": "nope"}), v))
}

func TestSet_WritesCanonicalName(t *testing.T) {
	v := config.Str("SET_NAME").WithAlias("SET_OLD_NAME")
	src := newSource(t, map[string]string{"SET_OLD_NAME": "old"})

	config.Set(src, v, "new")

	require.Equal(t, "new", config.Get(src, v))
	raw, ok, err := src.Provider().Get("SET_NAME")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "new", raw)
}

func TestWithAlias_KeepsNameCanonical(t *testing.T) {
	v := config.Str("ALIASES_NEW").WithAlias("ALIASES_OLD").WithAlias("ALIASES_OLDER")
	require.Equal(t, "ALIASES_NEW", v.Name())
	require.Equal(t, []string{"ALIASES_OLD", "ALIASES_OLDER"}, v.Aliases())
}

func newSource(t *testing.T, values map[string]string) *config.Source {
	t.Helper()

	src := config.NewSource(&config.MemProvider{}, nil)
	for name, val := range values {
		require.NoError(t, src.Provider().Set(name, val))
	}
	return src
}
