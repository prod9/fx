package files

import (
	"net/http"

	"fx.prodigy9.co/httpserver/controllers"
)

type (
	baseCtr struct {
		kind Kind
		mode Mode

		getOwnerID func(req *http.Request) int64
	}

	Option func(*baseCtr)
)

func Controller(kind Kind, options ...Option) controllers.Interface {
	if kind.Multiple {
		return multiFileCtr{baseCtr: newBaseCtr(kind, options...)}
	}
	return singleFileCtr{baseCtr: newBaseCtr(kind, options...)}
}

// newBaseCtr defaults to read-only: write access is the dangerous capability, so it
// must be opted into explicitly with WithMode(ModeReadWrite).
func newBaseCtr(kind Kind, options ...Option) baseCtr {
	b := baseCtr{
		kind:       kind,
		mode:       ModeReadOnly,
		getOwnerID: _getOwnerID,
	}
	for _, opt := range options {
		opt(&b)
	}
	return b
}

func WithMode(mode Mode) Option {
	return func(c *baseCtr) { c.mode = mode }
}
func WithOwnerIDFunc(getOwnerID func(req *http.Request) int64) Option {
	return func(c *baseCtr) { c.getOwnerID = getOwnerID }
}
