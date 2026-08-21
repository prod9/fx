package files

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/blobstore/blobserver"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

// TestRunCleanup drives the sweep end to end against a test database and a local
// blobserver: an abandoned row (no object, past dead) is pruned, while a healthy row
// (object present), a young row (upload still in flight), and a row older than the query
// window are all left in place.
func TestRunCleanup(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	createFilesTable(t, ctx)

	dir := t.TempDir()
	srv := httptest.NewServer(blobserver.NewHandler(dir))
	defer srv.Close()

	blobCfg := fxtest.Configure()
	storageURL := "http://key:secret@" + mustHost(t, srv.URL) + "/filesbucket"
	config.Set(blobCfg, blobstore.StorageURLConfig, storageURL)
	blobstore.DefaultClient = blobstore.NewClient(blobCfg)

	now := time.Now()
	abandoned := insertFile(t, ctx, 1, now.Add(-36*time.Hour))
	healthy := insertFile(t, ctx, 2, now.Add(-36*time.Hour))
	inFlight := insertFile(t, ctx, 3, now.Add(-1*time.Minute))
	preWindow := insertFile(t, ctx, 4, now.Add(-72*time.Hour))

	putObject(t, srv.URL, healthy.RemotePath(), "payload")

	require.NoError(t, runCleanup(ctx, now, 24*time.Hour, 500))

	require.False(t, fileExists(t, ctx, abandoned.ID), "abandoned upload must be pruned")
	require.True(t, fileExists(t, ctx, healthy.ID), "row with a live object must survive")
	require.True(t, fileExists(t, ctx, inFlight.ID), "in-flight upload must survive")
	require.True(t, fileExists(t, ctx, preWindow.ID), "row older than the window is not probed")
}

func TestFileFilterApply(t *testing.T) {
	createdAfter := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	createdBefore := createdAfter.Add(24 * time.Hour)
	filter := FileFilter{
		CreatedAfter:  createdAfter,
		CreatedBefore: createdBefore,
		AfterID:       41,
	}

	sql, args := filter.Apply("SELECT * FROM files", nil)

	require.Equal(t,
		"SELECT * FROM files WHERE created_at >= $1 AND created_at < $2 AND id > $3",
		sql,
	)
	require.Equal(t, []any{createdAfter, createdBefore, int64(41)}, args)
}

func createFilesTable(t *testing.T, ctx context.Context) {
	t.Helper()
	up, err := migrations.ReadFile("202504041033_create_files.up.sql")
	require.NoError(t, err)
	require.NoError(t, data.Exec(ctx, string(up)))
}

func insertFile(t *testing.T, ctx context.Context, id int64, createdAt time.Time) *File {
	t.Helper()
	f := &File{
		ID:            id,
		Kind:          "avatar",
		OwnerType:     "drop",
		OwnerID:       100,
		OriginalName:  "x",
		ContentType:   "text/plain",
		ContentLength: 1,
		CreatedAt:     createdAt,
	}
	require.NoError(t, data.Exec(ctx, `
		INSERT INTO files (
			id, kind, owner_id, owner_type, original_name,
			content_type, content_length, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		f.ID,
		f.Kind,
		f.OwnerID,
		f.OwnerType,
		f.OriginalName,
		f.ContentType,
		f.ContentLength,
		f.CreatedAt,
	))
	return f
}

func fileExists(t *testing.T, ctx context.Context, id int64) bool {
	t.Helper()
	var n int
	require.NoError(t, data.Get(ctx, &n, `SELECT COUNT(*) FROM files WHERE id = $1`, id))
	return n > 0
}

func putObject(t *testing.T, baseURL, key, body string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, baseURL+"/filesbucket/"+key, strings.NewReader(body))
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func mustHost(t *testing.T, raw string) string {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return u.Host
}
