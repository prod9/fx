package httpserver

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
}

func New(cfg *config.Source, mws []middlewares.Interface, ctrs []controllers.Interface) *Server {
	return &Server{cfg, mws, ctrs}
}

func (s *Server) Start() error {
	r := chi.NewRouter()
	for _, mws := range s.mws {
		r.Use(mws(s.cfg))
	}
	for _, ctr := range s.ctrs {
		if err := ctr.Mount(s.cfg, r); err != nil {
			return err
		}
	}
	r.NotFound(func(resp http.ResponseWriter, req *http.Request) {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
	})

	listenAddr := config.Get(s.cfg, ListenAddrConfig)
	srv := http.Server{
		Addr: listenAddr,
	}

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctrlC
		srv.Shutdown(nil)
	}()

	log.Println("listening on " + listenAddr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	} else {
		return err
	}
}
