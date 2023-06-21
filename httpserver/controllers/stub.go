package controllers

import (
	"fx.prodigy9.co/config"
	"github.com/go-chi/chi/v5"
)

type Stub struct {
	MountFunc func(cfg *config.Source, router chi.Router) error
}

var _ Interface = Stub{}

func (s Stub) Mount(cfg *config.Source, router chi.Router) error {
	return s.MountFunc(cfg, router)
}
