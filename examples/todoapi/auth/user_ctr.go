package auth

import (
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

type UserCtr struct{}

func (c UserCtr) Mount(cfg *config.Source, router chi.Router) error {
	router.Post("/users", c.Create)
	router.Group(func(router chi.Router) {
		router.Use(RequireSession(cfg))
		router.Get("/users/current", c.Current)
	})
	return nil
}

func (c UserCtr) Create(resp http.ResponseWriter, req *http.Request) {
	action, user := &CreateUser{}, &User{}
	if err := controllers.ExecuteAction(resp, req, action, user); err != nil {
		render.Error(resp, req, 400, err)
	} else {
		render.JSON(resp, req, user)
	}
}

func (c UserCtr) Current(resp http.ResponseWriter, req *http.Request) {
	user := UserFromContext(req.Context())
	if user == nil {
		render.Error(resp, req, 403, httperrors.ErrUnauthorized)
	} else {
		render.JSON(resp, req, user)
	}
}
