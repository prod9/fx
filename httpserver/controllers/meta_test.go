package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fx.prodigy9.co/buildinfo"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func mountMeta(t *testing.T, cfg *config.Source) chi.Router {
	t.Helper()
	r := chi.NewRouter()
	require.NoError(t, Meta("/.nx").Mount(cfg, r))
	return r
}

func doGet(r chi.Router, ctx context.Context, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestMeta_Index(t *testing.T) {
	w := doGet(mountMeta(t, fxtest.Configure()), context.Background(), "/")
	require.Equal(t, http.StatusOK, w.Code)
}

func TestMeta_Version(t *testing.T) {
	ctx := buildinfo.NewContext(context.Background(), buildinfo.Info{Name: "api", Version: "v1.2.3"})

	w := doGet(mountMeta(t, fxtest.Configure()), ctx, "/.nx/version")
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "v1.2.3", w.Body.String())
}

func TestMeta_BuildInfo(t *testing.T) {
	want := buildinfo.Info{Name: "api", Version: "v1.2.3", Revision: "abc123"}
	ctx := buildinfo.NewContext(context.Background(), want)

	w := doGet(mountMeta(t, fxtest.Configure()), ctx, "/.nx/buildinfo")
	require.Equal(t, http.StatusOK, w.Code)
	var got buildinfo.Info
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	require.Equal(t, want, got)
}

func TestMeta_PanicGatedByDebugSentry(t *testing.T) {
	off := doGet(mountMeta(t, fxtest.Configure()), context.Background(), "/.nx/panic")
	require.Equal(t, http.StatusNotFound, off.Code)

	cfg := fxtest.Configure()
	config.Set(cfg, DebugSentryConfig, true)
	r := mountMeta(t, cfg)
	require.Panics(t, func() { doGet(r, context.Background(), "/.nx/panic") })
}

func TestHealthz_NoDB(t *testing.T) {
	w := doGet(mountMeta(t, fxtest.Configure()), context.Background(), "/healthz")

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Equal(t, "ok", body["status"])
	_, hasDB := body["db"]
	require.False(t, hasDB, "db field should be absent when no DB in context")
}

func TestHealthz_PingOK(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	w := doGet(mountMeta(t, fxtest.Configure()), ctx, "/healthz")

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Equal(t, "ok", body["status"])
	require.Equal(t, "ok", body["db"])
}

func TestHealthz_PingFails(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	db := data.FromContext(ctx)
	require.NoError(t, db.Close())

	w := doGet(mountMeta(t, fxtest.Configure()), ctx, "/healthz")
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
}
