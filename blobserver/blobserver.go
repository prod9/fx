package blobserver

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/ctrlc"
	"fx.prodigy9.co/fxlog"
)

var (
	ListenAddrConfig = config.StrDef("BLOBSERVER_ADDR", "0.0.0.0:9500")
	StorageDirConfig = config.StrDef("BLOBSERVER_DIR", "tmp/blobstore")
)

type Server struct {
	cfg *config.Source
}

func New(cfg *config.Source) *Server {
	return &Server{cfg}
}

func (s *Server) Start() error {
	dir, err := filepath.Abs(config.Get(s.cfg, StorageDirConfig))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	listenAddr := config.Get(s.cfg, ListenAddrConfig)
	srv := http.Server{
		Addr:    listenAddr,
		Handler: NewHandler(dir),
	}
	ctrlc.Do(func() { srv.Shutdown(context.Background()) })

	fxlog.Log("blob serving",
		fxlog.String("addr", listenAddr),
		fxlog.String("dir", dir))

	err = srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
