package httpserver

import (
	"log"
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/ctrlc"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/middlewares"
	"github.com/go-chi/chi/v5"
)

var ListenAddrConfig = config.StrDef("LISTEN_ADDR", "0.0.0.0:3000")

type Server struct {
	cfg       *config.Source
	fragments []*Fragment
}

func New(cfg *config.Source, mws []middlewares.Interface, ctrs []controllers.Interface) *Server {
	return &Server{cfg, []*Fragment{NewFragment(mws, ctrs)}}
}
func NewWithFragments(cfg *config.Source, fragments []*Fragment) *Server {
	return &Server{cfg, fragments}
}

func (s *Server) Start() error {
	router := chi.NewRouter()

	for _, frag := range s.fragments {
		if err := frag.configureRoutes(s.cfg, router); err != nil {
			return err
		}
	}

	listenAddr := config.Get(s.cfg, ListenAddrConfig)
	srv := http.Server{
		Addr:    listenAddr,
		Handler: router,
	}

	ctrlc.Do(func() {
		srv.Shutdown(nil)
	})

	log.Println("listening on " + listenAddr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	} else {
		return err
	}
}
