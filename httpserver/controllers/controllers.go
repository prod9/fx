package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

type (
	Interface interface {
		Mount(cfg *config.Source, router chi.Router) error
	}

	// standalone action
	Validator interface {
		Validate() error
	}
	Action interface {
		Execute(ctx context.Context, out any) error
	}
)

func ReadJSON(r *http.Request, obj interface{}) error {
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		return fmt.Errorf("read json: %w", err)
	}
	return nil
}

func ReadAction(r *http.Request, act Action) error {
	if err := ReadJSON(r, act); err != nil {
		return err
	}

	if val, ok := act.(Validator); ok {
		if err := val.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func ExecuteAction(resp http.ResponseWriter, r *http.Request, act Action, out any) error {
	if err := ReadAction(r, act); err != nil {
		return err
	} else if err := act.Execute(r.Context(), out); err != nil {
		return err
	} else {
		return nil
	}
}

// Static creates a controllers.Interface that simply renders the given object as JSON on
// an incoming HTTP GET.
func StaticJSON(path string, obj any) Interface {
	return Stub{
		MountFunc: func(_ *config.Source, router chi.Router) error {
			router.Get(path, func(resp http.ResponseWriter, r *http.Request) {
				render.JSON(resp, r, obj)
			})
			return nil
		},
	}
}

func FromFunc(path string, f http.HandlerFunc) Interface {
	return Stub{
		MountFunc: func(_ *config.Source, router chi.Router) error {
			router.HandleFunc(path, f)
			return nil
		},
	}
}

func FromHandler(path string, h http.Handler) Interface {
	return Stub{
		MountFunc: func(_ *config.Source, router chi.Router) error {
			router.Handle(path, h)
			return nil
		},
	}
}
