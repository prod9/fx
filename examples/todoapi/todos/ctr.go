package todos

import (
	"net/http"
	"strconv"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/examples/todoapi/auth"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

type Ctr struct{}

func (c Ctr) Mount(cfg *config.Source, router chi.Router) error {
	router.Group(func(router chi.Router) {
		router.Use(auth.RequireSession(cfg))

		router.Get("/todos", c.Index)
		router.Post("/todos", c.Create)
		router.Patch("/todos/{id}", c.Update)
		router.Delete("/todos/{id}", c.Delete)
	})
	return nil
}

func (c Ctr) Index(resp http.ResponseWriter, req *http.Request) {
	user := auth.UserFromContext(req.Context())
	if todos, err := GetTodosByUserID(req.Context(), user.ID); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, todos)
	}
}

func (c Ctr) Create(resp http.ResponseWriter, req *http.Request) {
	action, todo := &CreateTodo{}, &Todo{}
	if err := controllers.ExecuteAction(resp, req, action, todo); err != nil {
		render.Error(resp, req, 400, err)
	} else {
		render.JSON(resp, req, todo)
	}
}

func (c Ctr) Update(resp http.ResponseWriter, req *http.Request) {
	id := c.getID(req)
	if id < 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	action, todo := &UpdateTodo{ID: id}, &Todo{}
	if err := controllers.ExecuteAction(resp, req, action, todo); err != nil {
		render.Error(resp, req, 400, err)
	} else {
		render.JSON(resp, req, todo)
	}
}

func (c Ctr) Delete(resp http.ResponseWriter, req *http.Request) {
	id := c.getID(req)
	if id < 0 {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
		return
	}

	user := auth.UserFromContext(req.Context())
	if todo, err := DeleteTodo(req.Context(), user.ID, id); err != nil {
		render.Error(resp, req, 400, err)
	} else {
		render.JSON(resp, req, todo)
	}
}

func (c Ctr) getID(req *http.Request) int64 {
	raw := chi.URLParam(req, "id")
	if id, err := strconv.ParseInt(raw, 10, 64); err != nil {
		return -1
	} else {
		return id
	}
}
