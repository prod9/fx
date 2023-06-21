package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"fx.prodigy9.co/config"
	"github.com/go-chi/chi/v5"
)

var (
	ErrNotFound     = &errImpl{"not_found", "not found"}
	ErrUnauthorized = &errImpl{"unauthorized", "unauthorized"}
	ErrInternal     = &errImpl{"internal", "internal server error"}
	ErrBadRequest   = &errImpl{"bad_request", "bad request"}
)

type errImpl struct {
	code    string
	message string
}

func (i *errImpl) Code() string  { return i.code }
func (i *errImpl) Error() string { return i.message }

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
