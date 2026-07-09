package drops

import (
	"net/http"

	"fx.prodigy9.co/app/files"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

// fileKind ties each drop to a single stored file. A drop accepts common document
// and image types plus arbitrary binaries via application/octet-stream.
var fileKind = files.Kind{
	Name:      "drop",
	OwnerType: "drop",
	ContentTypes: []string{
		"application/octet-stream",
		"application/pdf",
		"application/zip",
		"image/jpeg",
		"image/png",
		"text/plain",
	},
}

type Ctr struct{}

var _ controllers.Interface = Ctr{}

func (c Ctr) Mount(cfg *config.Source, router chi.Router) error {
	router.Post("/drops", c.Create)

	// The files controller resolves the owner (the drop) from the URL token, then
	// exposes POST/GET/DELETE for that drop's single file under /d/{token}/file.
	// Writes are opt-in (the controller defaults to read-only).
	fileCtr := files.Controller(fileKind,
		files.WithMode(files.ModeReadWrite),
		files.WithOwnerIDFunc(ownerIDFromToken))

	var mountErr error
	router.Route("/d/{token}/file", func(r chi.Router) {
		mountErr = fileCtr.Mount(cfg, r)
	})
	return mountErr
}

func (c Ctr) Create(resp http.ResponseWriter, req *http.Request) {
	action, drop := &CreateDrop{}, &Drop{}
	if err := action.Execute(req.Context(), drop); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, drop)
	}
}

func ownerIDFromToken(req *http.Request) int64 {
	token := chi.URLParam(req, "token")
	if token == "" {
		return 0
	}

	drop, err := GetDropByToken(req.Context(), token)
	if err != nil {
		return 0
	}
	return drop.ID
}
