package settings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// TestMountServesSettings drives the settings endpoints end to end through the exported
// MountRoutes entrypoint: an empty list is served, an upsert creates a row, and a follow-up
// list reflects it. It also stands in for the mount-auth contract — routes exist only where
// the consumer calls MountRoutes, never from the fragment itself.
func TestMountServesSettings(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	createSettingsTable(t, ctx)
	cfg := config.FromContext(ctx)
	db := data.FromContext(ctx)

	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := config.NewContext(req.Context(), cfg)
			c = data.NewContext(c, db)
			next.ServeHTTP(w, req.WithContext(c))
		})
	})
	require.NoError(t, MountRoutes(cfg, router))

	srv := httptest.NewServer(router)
	defer srv.Close()

	require.Empty(t, listSettings(t, srv.URL), "a fresh table must serve an empty list")

	created := postSetting(t, srv.URL, "theme", "dark")
	require.Equal(t, "theme", created.Key)
	require.Equal(t, "dark", created.Value)

	list := listSettings(t, srv.URL)
	require.Len(t, list, 1)
	require.Equal(t, "dark", list[0].Value)
}

func listSettings(t *testing.T, baseURL string) []*Settings {
	t.Helper()
	resp, err := http.Get(baseURL + "/settings")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out []*Settings
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

func postSetting(t *testing.T, baseURL, key, value string) *Settings {
	t.Helper()
	body := strings.NewReader(`{"value":"` + value + `"}`)
	resp, err := http.Post(baseURL+"/settings/"+key, "application/json", body)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	out := &Settings{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(out))
	return out
}
