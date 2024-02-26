package controllers

import (
	"context"
	"net/http"
)

type (
	Validator interface {
		Validate() error
	}
	Action interface {
		Execute(ctx context.Context, out any) error
	}
)

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
	}
	if err := act.Execute(r.Context(), out); err != nil {
		return err
	} else {
		return nil
	}
}
