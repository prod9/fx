package files

import (
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/page"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

type multiFileCtr struct {
	baseCtr
}

var _ controllers.Interface = multiFileCtr{}

func (f multiFileCtr) Mount(cfg *config.Source, r chi.Router) error {
	if f.mode&modeRead > 0 {
		r.Get("/", f.List)
		r.Get("/{fileID}", f.Get)
		r.Get("/{fileID}/meta", f.GetMeta)
	}
	if f.mode&modeWrite > 0 {
		r.Post("/", f.Create)
		r.Delete("/{fileID}", f.Destroy)
	}
	return nil
}

func (f multiFileCtr) List(resp http.ResponseWriter, req *http.Request) {
	ownerID := f.getOwnerID(req)
	if ownerID <= 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	key, pm := f.kind.key(ownerID, 0), page.FromRequest(req)
	if files, err := ListFiles(req.Context(), key, pm); err != nil {
		if data.IsNoRows(err) {
			render.JSON(resp, req, page.Empty[*File]())
		} else {
			render.Error(resp, req, 500, httperrors.ErrInternal)
		}
	} else {
		render.JSON(resp, req, files)
	}
}

func (f multiFileCtr) Get(resp http.ResponseWriter, req *http.Request) {
	ownerID, fileID := f.getOwnerID(req), getFileID(req)
	if ownerID <= 0 || fileID <= 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	key := f.kind.key(ownerID, fileID)
	if file, err := GetFileByID(req.Context(), key); err != nil {
		if data.IsNoRows(err) {
			render.Error(resp, req, 404, httperrors.ErrNotFound)
		} else {
			render.Error(resp, req, 500, err)
		}
	} else if url, err := file.PresignedGetURL(req.Context()); err != nil {
		render.Error(resp, req, 500, httperrors.ErrInternal)
	} else {
		render.Redirect(resp, req, url)
	}
}

func (f multiFileCtr) GetMeta(resp http.ResponseWriter, req *http.Request) {
	ownerID, fileID := f.getOwnerID(req), getFileID(req)
	if ownerID <= 0 || fileID <= 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	key := f.kind.key(ownerID, fileID)
	if file, err := GetFileByID(req.Context(), key); err != nil {
		if data.IsNoRows(err) {
			render.Error(resp, req, 404, httperrors.ErrNotFound)
		} else {
			render.Error(resp, req, 500, err)
		}
	} else {
		render.JSON(resp, req, file)
	}
}

func (f multiFileCtr) Create(resp http.ResponseWriter, req *http.Request) {
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
	} else if info, err := UploadInfoFromFile(req.Context(), file); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, info)
	}
}

func (f multiFileCtr) Destroy(resp http.ResponseWriter, req *http.Request) {
	ownerID, fileID := f.getOwnerID(req), getFileID(req)
	if ownerID <= 0 || fileID <= 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	key := f.kind.key(ownerID, fileID)
	if file, err := DestroyFile(req.Context(), key); err != nil {
		if data.IsNoRows(err) {
			render.Error(resp, req, 404, httperrors.ErrNotFound)
		} else {
			render.Error(resp, req, 500, err)
		}
	} else {
		render.JSON(resp, req, file)
	}
}
