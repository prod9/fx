package blobserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func serve(h http.Handler, method, target string, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, reader))
	return rec
}

func TestHandlerPutThenGet(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(dir)
	body := "hello blob"

	put := serve(h, http.MethodPut, "/bucket/path/to/obj.txt", body)
	require.Equal(t, http.StatusOK, put.Code)

	onDisk, err := os.ReadFile(filepath.Join(dir, "bucket", "path", "to", "obj.txt"))
	require.NoError(t, err)
	require.Equal(t, body, string(onDisk))

	get := serve(h, http.MethodGet, "/bucket/path/to/obj.txt", "")
	require.Equal(t, http.StatusOK, get.Code)
	require.Equal(t, body, get.Body.String())
}

func TestHandlerGetMissing(t *testing.T) {
	h := NewHandler(t.TempDir())
	get := serve(h, http.MethodGet, "/bucket/nope.txt", "")
	require.Equal(t, http.StatusNotFound, get.Code)
}

func TestHandlerDelete(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(dir)

	serve(h, http.MethodPut, "/bucket/x.txt", "data")

	del := serve(h, http.MethodDelete, "/bucket/x.txt", "")
	require.Equal(t, http.StatusNoContent, del.Code)

	get := serve(h, http.MethodGet, "/bucket/x.txt", "")
	require.Equal(t, http.StatusNotFound, get.Code)
}

func TestHandlerNeutralizesTraversal(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(dir)

	serve(h, http.MethodPut, "/bucket/../../escape.txt", "x")

	_, err := os.Stat(filepath.Join(filepath.Dir(dir), "escape.txt"))
	require.True(t, os.IsNotExist(err), "traversal must not write outside the root")
}
