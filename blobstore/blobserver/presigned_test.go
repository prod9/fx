package blobserver_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/blobstore/blobserver"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

// Exercises the full local-dev path: a blobstore.Client pointed at a plaintext
// blobserver (http:// scheme) mints presigned URLs that drive the server end to end.
func TestPresignedRoundTripAgainstBlobserver(t *testing.T) {
	dir := t.TempDir()
	srv := httptest.NewServer(blobserver.NewHandler(dir))
	defer srv.Close()

	host := must(url.Parse(srv.URL)).Host
	cfg := fxtest.Configure()
	config.Set(cfg, blobstore.StorageURLConfig, "http://key:secret@"+host+"/testbucket")
	client := blobstore.NewClient(cfg)

	ctx := context.Background()
	key, body := "hello/world.txt", "presigned payload"

	putURL, err := client.PresignedPutURL(ctx, key)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(putURL, "http://"), "local presign must be plaintext")
	putResp, err := http.DefaultClient.Do(must(http.NewRequest(http.MethodPut, putURL, strings.NewReader(body))))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, putResp.StatusCode)

	onDisk, err := os.ReadFile(filepath.Join(dir, "testbucket", key))
	require.NoError(t, err)
	require.Equal(t, body, string(onDisk))

	getURL, err := client.PresignedGetURL(ctx, key)
	require.NoError(t, err)
	getResp, err := http.Get(getURL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getResp.StatusCode)
	got, _ := io.ReadAll(getResp.Body)
	require.Equal(t, body, string(got))

	require.NoError(t, client.DeleteObject(ctx, key))
	_, err = os.Stat(filepath.Join(dir, "testbucket", key))
	require.True(t, os.IsNotExist(err), "delete must remove the file")
}

// Gates that minio-go's ListObjects can parse blobserver's ListObjectsV2 XML — the
// reconciliation sweep depends on it.
func TestListObjectsAgainstBlobserver(t *testing.T) {
	dir := t.TempDir()
	srv := httptest.NewServer(blobserver.NewHandler(dir))
	defer srv.Close()

	host := must(url.Parse(srv.URL)).Host
	cfg := fxtest.Configure()
	config.Set(cfg, blobstore.StorageURLConfig, "http://key:secret@"+host+"/testbucket")
	client := blobstore.NewClient(cfg)

	ctx := context.Background()
	for _, key := range []string{"a/one.txt", "b/two.txt"} {
		putURL, err := client.PresignedPutURL(ctx, key)
		require.NoError(t, err)
		_, err = http.DefaultClient.Do(must(http.NewRequest(http.MethodPut, putURL, strings.NewReader("x"))))
		require.NoError(t, err)
	}

	keys, err := client.ListObjects(ctx, "")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"a/one.txt", "b/two.txt"}, keys)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
