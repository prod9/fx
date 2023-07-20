package httpserver

import (
	"log"
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/httpserver/middlewares"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

var ListenAddrConfig = config.StrDef("LISTEN_ADDR", "0.0.0.0:3000")

type Server struct {
	cfg  *config.Source
	mws  []middlewares.Interface
	ctrs []controllers.Interface
	r    *chi.Mux
}

func New(cfg *config.Source, mws []middlewares.Interface, ctrs []controllers.Interface) *Server {
	return &Server{cfg, mws, ctrs, chi.NewRouter()}
}

func (s *Server) GetRouter() *chi.Mux {
	return s.r
}

func (s *Server) PrepareRouter() error {
	for _, mws := range s.mws {
		s.r.Use(mws(s.cfg))
	}
	for _, ctr := range s.ctrs {
		if err := ctr.Mount(s.cfg, s.r); err != nil {
			return err
		}
	}
	s.r.NotFound(func(resp http.ResponseWriter, req *http.Request) {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
	})
	return nil
}

func (s *Server) Start() error {
	listenAddr := config.Get(s.cfg, ListenAddrConfig)
	log.Println("listening on " + listenAddr)
	return http.ListenAndServe(listenAddr, s.r)
}
