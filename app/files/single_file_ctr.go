package files

import (
	"net/http"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

type singleFileCtr struct {
	baseCtr
	linkAgeCfg time.Duration // resolved from config at mount time
}

var _ controllers.Interface = singleFileCtr{}

func (f singleFileCtr) Mount(cfg *config.Source, r chi.Router) error {
	f.linkAgeCfg = f.resolveLinkAge(cfg)

	if f.mode&modeRead > 0 {
		r.Get("/", f.Get)
		r.Get("/meta", f.GetMeta)
	}
	if f.mode&modeWrite > 0 {
		r.Post("/", f.Create)
		r.Delete("/", f.Destroy)
	}
	return nil
}

func (f singleFileCtr) Get(resp http.ResponseWriter, req *http.Request) {
	ownerID := f.getOwnerID(req)
	if ownerID <= 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	key := f.kind.key(ownerID, 0)
	if file, err := GetUniqueFile(req.Context(), key); err != nil {
		if data.IsNoRows(err) {
			render.Error(resp, req, 404, httperrors.ErrNotFound)
		} else {
			render.Error(resp, req, 500, err)
		}
	} else if url, err := file.PresignedGetURL(req.Context(), f.client, f.linkAgeCfg); err != nil {
		render.Error(resp, req, 500, httperrors.ErrInternal)
	} else {
		render.Redirect(resp, req, url)
	}
}

func (f singleFileCtr) GetMeta(resp http.ResponseWriter, req *http.Request) {
	ownerID := f.getOwnerID(req)
	if ownerID <= 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	key := f.kind.key(ownerID, 0)
	if file, err := GetUniqueFile(req.Context(), key); err != nil {
		if data.IsNoRows(err) {
			render.Error(resp, req, 404, httperrors.ErrNotFound)
		} else {
			render.Error(resp, req, 500, err)
		}
	} else {
		render.JSON(resp, req, file)
	}
}

func (f singleFileCtr) Create(resp http.ResponseWriter, req *http.Request) {
	ownerID := f.getOwnerID(req)
	if ownerID <= 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	action, file := &CreateFile{
		Kind:    f.kind,
		OwnerID: ownerID,
	}, &File{}
	if err := controllers.ExecuteAction(resp, req, action, file); err != nil {
		render.Error(resp, req, 400, err)
	} else if info, err := UploadInfoFromFile(req.Context(), f.client, file, f.linkAgeCfg); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, info)
	}
}

func (f singleFileCtr) Destroy(resp http.ResponseWriter, req *http.Request) {
	ownerID := f.getOwnerID(req)
	if ownerID <= 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	key := f.kind.key(ownerID, 0)
	if file, err := DestroyUniqueFile(req.Context(), f.client, key); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, file)
	}
}
