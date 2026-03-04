package files

import (
	"embed"
	"net/http"
	"strconv"
	"time"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/config"
	"github.com/go-chi/chi/v5"
)

//go:embed *.sql
var migrations embed.FS

// LinkAgeConfig is the app-wide default for presigned URL expiry.
// Override per-controller with WithLinkAge option.
var LinkAgeConfig = config.DurationDef("FILE_LINK_AGE", 1*time.Minute)

// App is the default files app fragment using the global blobstore functions.
var App = NewApp(nil)

// NewApp creates a files app fragment with an optional specific blobstore client.
// Pass nil to use the global blobstore functions.
func NewApp(client *blobstore.Client) *app.Builder {
	_ = client // stored for reference; controllers receive client via WithClient option
	return app.Build().
		EmbedMigrations(migrations)
}

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

func _getOwnerID(req *http.Request) int64 {
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
