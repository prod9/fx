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

func TestHandlerList(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(dir)
	serve(h, http.MethodPut, "/bucket/a.txt", "a")
	serve(h, http.MethodPut, "/bucket/sub/b.txt", "bb")
	serve(h, http.MethodPut, "/other/c.txt", "ccc")

	all := serve(h, http.MethodGet, "/bucket?list-type=2", "")
	require.Equal(t, http.StatusOK, all.Code)
	require.Contains(t, all.Body.String(), "<Key>a.txt</Key>")
	require.Contains(t, all.Body.String(), "<Key>sub/b.txt</Key>")
	require.NotContains(t, all.Body.String(), "c.txt", "other bucket must not leak")

	pref := serve(h, http.MethodGet, "/bucket?list-type=2&prefix=sub/", "")
	require.Contains(t, pref.Body.String(), "<Key>sub/b.txt</Key>")
	require.NotContains(t, pref.Body.String(), "<Key>a.txt</Key>")
}

func TestHandlerListMissingBucket(t *testing.T) {
	h := NewHandler(t.TempDir())
	rec := serve(h, http.MethodGet, "/nope?list-type=2", "")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "<KeyCount>0</KeyCount>")
}

func TestHandlerNeutralizesTraversal(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(dir)

	serve(h, http.MethodPut, "/bucket/../../escape.txt", "x")

	_, err := os.Stat(filepath.Join(filepath.Dir(dir), "escape.txt"))
	require.True(t, os.IsNotExist(err), "traversal must not write outside the root")
}
