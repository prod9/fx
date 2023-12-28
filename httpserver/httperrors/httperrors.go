package httperrors

import "fx.prodigy9.co/errutil"

var (
	ErrNotFound     = errutil.NewCoded("not_found", "not found", nil)
	ErrUnauthorized = errutil.NewCoded("unauthorized", "unauthorized", nil)
	ErrInternal     = errutil.NewCoded("internal", "internal server error", nil)
	ErrBadRequest   = errutil.NewCoded("bad_request", "bad request", nil)
)
