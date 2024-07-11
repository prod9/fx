package httpserver

import (
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/httpserver/middlewares"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

// Fragment encapsulates a set of middlewares and controllers allowing parts of an HTTP
// application to be split up into reusable modules that works independently of one
// another (for example, using a very different set of middlewares) without requiring
// per-controller configuration.
type Fragment struct {
	mws  []middlewares.Interface
	ctrs []controllers.Interface

	children []*Fragment
}

func NewFragment(mws []middlewares.Interface, ctrs []controllers.Interface) *Fragment {
	return &Fragment{mws, ctrs, nil}
}

func (f *Fragment) IsEmpty() bool {
	return f == nil ||
		(len(f.mws) == 0 &&
			len(f.ctrs) == 0)
}
func (f *Fragment) HasNoMiddlewares() bool {
	return f == nil || (len(f.mws) == 0)
}

func (f *Fragment) AddChild(fragment *Fragment) *Fragment {
	f.children = append(f.children, fragment)
	return f
}
func (f *Fragment) AddMiddlewares(mws ...middlewares.Interface) *Fragment {
	f.mws = append(f.mws, mws...)
	return f
}
func (f *Fragment) AddControllers(ctrs ...controllers.Interface) *Fragment {
	f.ctrs = append(f.ctrs, ctrs...)
	return f
}

func (f *Fragment) configureRoutes(cfg *config.Source, router chi.Router) error {
	for _, mws := range f.mws {
		router.Use(mws(cfg))
	}
	for _, ctr := range f.ctrs {
		if err := ctr.Mount(cfg, router); err != nil {
			return err
		}
	}

	// mount child routers using Group (so we get independent middlewares stack)
	var innerErr error = nil
	for _, child := range f.children {
		router.Group(func(r chi.Router) {
			innerErr = child.configureRoutes(cfg, r)
		})
		if innerErr != nil {
			return innerErr
		}
	}

	router.NotFound(func(resp http.ResponseWriter, req *http.Request) {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
	})
	return nil
}
