package auth

import (
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

type SessionCtr struct{}

func (c SessionCtr) Mount(cfg *config.Source, router chi.Router) error {
	router.Post("/sessions", c.Create)
	router.Delete("/sessions/current", c.Destroy)

	router.Group(func(router chi.Router) {
		router.Use(RequireSession(cfg))
		router.Get("/sessions/current", c.Current)
	})
	return nil
}

func (c SessionCtr) Create(resp http.ResponseWriter, req *http.Request) {
	action, sess := &CreateSession{}, &Session{}
	if err := controllers.ExecuteAction(resp, req, action, sess); err != nil {
		render.Error(resp, req, 400, err)
	} else {
		render.JSON(resp, req, sess)
	}
}

func (c SessionCtr) Current(resp http.ResponseWriter, req *http.Request) {
	sess := SessionFromContext(req.Context())
	if sess == nil {
		render.Error(resp, req, 403, httperrors.ErrUnauthorized)
	} else {
		render.JSON(resp, req, sess)
	}
}

func (c SessionCtr) Destroy(resp http.ResponseWriter, req *http.Request) {
	token := GetBearerTokenFromRequest(req) // no need to send full JSON payload
	action := &DestroySession{Token: token}
	if err := action.Validate(); err != nil {
		render.Error(resp, req, 400, err)
	}

	sess := &Session{}
	if err := action.Execute(req.Context(), sess); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, sess)
	}
}
