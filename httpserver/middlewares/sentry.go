package middlewares

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"fx.prodigy9.co/config"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

var (
	SentryDSNConfig = config.Str("API_SENTRY_DSN")
)

type sentryContainer struct {
	sync.RWMutex

	err     error
	hub     *sentry.Hub
	handler *sentryhttp.Handler
}

func (c *sentryContainer) get(cfg *config.Source) (*sentry.Hub, *sentryhttp.Handler) {
	if err := c.tryGetErr(); err != nil {
		log.Printf("sentry: %s", err)
		return nil, nil
	} else if client, handler := c.tryGet(); client != nil {
		return client, handler
	}

	c.initialize(cfg)
	return c.get(cfg)
}

func (c *sentryContainer) tryGetErr() error {
	c.RLock()
	defer c.RUnlock()
	return c.err
}

func (c *sentryContainer) tryGet() (*sentry.Hub, *sentryhttp.Handler) {
	c.RLock()
	defer c.RUnlock()
	return c.hub, c.handler
}

func (c *sentryContainer) initialize(cfg *config.Source) {
	c.Lock()
	defer c.Unlock()

	dsn := config.Get(cfg, SentryDSNConfig)
	if dsn == "" {
		c.err = errors.New("API_SENTRY_DSN not configured")
		return
	}

	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: dsn})
	if err != nil {
		c.err = err
		return
	}

	hub := sentry.NewHub(client, sentry.NewScope())
	handler := sentryhttp.New(sentryhttp.Options{Repanic: true})
	c.hub, c.handler = hub, handler
	return
}

func Sentry(cfg *config.Source) func(http.Handler) http.Handler {
	container := &sentryContainer{}

	return func(handler http.Handler) http.Handler {
		sentryHub, sentryHandler := container.get(cfg)
		if sentryHub == nil || sentryHandler == nil {
			log.Println("sentry: not setup")
			return handler
		}

		handler = sentryHandler.Handle(handler)

		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			var (
				ctx = req.Context()
				hub = sentry.GetHubFromContext(ctx)
			)
			if hub == nil {
				hub = sentryHub.Clone()
				ctx = sentry.SetHubOnContext(ctx, hub)
			}

			handler.ServeHTTP(resp, req.WithContext(ctx))
		})
	}
}
