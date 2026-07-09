package files

import (
	"embed"
	"net/http"
	"strconv"
	"time"

	"fx.prodigy9.co/app"
	"github.com/go-chi/chi/v5"
)

//go:embed *.sql
var migrations embed.FS

// linkAge is the presigned URL TTL. Access is gated one level up (the route that
// embeds the files controller), so the TTL is a fixed policy, not a per-request knob.
const linkAge = 1 * time.Minute

// App is the files fragment: the files metadata table plus the reconciliation worker
// that prunes orphaned rows and objects. Mount it, then embed files.Controller(...)
// into your own routes to expose upload/download endpoints.
var App = app.Build().
	Name("files").
	EmbedMigrations(migrations).
	Job(filesCleanup)

var ImageTypes = []string{
	"image/jpeg",
	"image/png",
	"image/webp",
}

type Mode uint8

const (
	modeRead = 1 << iota
	modeWrite

	ModeReadOnly  = modeRead
	ModeReadWrite = modeRead | modeWrite
)

func defaultOwnerID(req *http.Request) int64 {
	if id_ := chi.URLParam(req, "id"); id_ == "" {
		return 0
	} else if id, err := strconv.ParseInt(id_, 10, 64); err != nil {
		return 0
	} else {
		return id
	}
}

func getFileID(req *http.Request) int64 {
	if id_ := chi.URLParam(req, "fileID"); id_ == "" {
		return 0
	} else if id, err := strconv.ParseInt(id_, 10, 64); err != nil {
		return 0
	} else {
		return id
	}
}
