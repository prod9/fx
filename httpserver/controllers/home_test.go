package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func mountHome(t *testing.T) chi.Router {
	t.Helper()
	r := chi.NewRouter()
	require.NoError(t, Home{}.Mount(nil, r))
	return r
}

func doGet(r chi.Router, ctx context.Context, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealthz_NoDB(t *testing.T) {
	w := doGet(mountHome(t), context.Background(), "/healthz")

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	require.Equal(t, "ok", body["status"])
	_, hasDB := body["db"]
	require.False(t, hasDB, "db field should be absent when no DB in context")
}

func TestHealthz_PingOK(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	w := doGet(mountHome(t), ctx, "/healthz")

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

	w := doGet(mountHome(t), ctx, "/healthz")
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
}
