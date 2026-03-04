package files

import (
	"net/http"
	"time"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver/controllers"
)

type (
	baseCtr struct {
		kind Kind
		mode Mode

		client  *blobstore.Client
		linkAge time.Duration // 0 means use LinkAgeConfig

		getOwnerID func(req *http.Request) int64
	}

	Option func(*baseCtr)
)

func Controller(kind Kind, options ...Option) controllers.Interface {
	if kind.Multiple {
		return multiFileCtr{baseCtr: newBaseCtr(kind, options...)}
	} else {
		return singleFileCtr{baseCtr: newBaseCtr(kind, options...)}
	}
}

func newBaseCtr(kind Kind, options ...Option) baseCtr {
	b := baseCtr{
		kind:       kind,
		mode:       ModeReadWrite,
		getOwnerID: _getOwnerID,
	}
	for _, opt := range options {
		opt(&b)
	}
	return b
}

// resolveLinkAge returns the effective link age for this controller.
// Resolution order: WithLinkAge option -> FILE_LINK_AGE env var -> 1 minute default.
func (b *baseCtr) resolveLinkAge(cfg *config.Source) time.Duration {
	if b.linkAge > 0 {
		return b.linkAge
	}
	return config.Get(cfg, LinkAgeConfig)
}

func WithKind(kind Kind) Option {
	return func(c *baseCtr) { c.kind = kind }
}
func WithMode(mode Mode) Option {
	return func(c *baseCtr) { c.mode = mode }
}
func WithOwnerIDFunc(getOwnerID func(req *http.Request) int64) Option {
	return func(c *baseCtr) { c.getOwnerID = getOwnerID }
}
func WithClient(client *blobstore.Client) Option {
	return func(c *baseCtr) { c.client = client }
}
func WithLinkAge(age time.Duration) Option {
	return func(c *baseCtr) { c.linkAge = age }
}
