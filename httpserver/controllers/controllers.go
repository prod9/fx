package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"fx.prodigy9.co/config"
	"github.com/go-chi/chi/v5"
)

type Interface interface {
	Mount(cfg *config.Source, router chi.Router) error
}

func ReadJSON(r *http.Request, obj interface{}) error {
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		return fmt.Errorf("read json: %w", err)
	}
	return nil
}
