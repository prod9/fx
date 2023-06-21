package middlewares

import (
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/httpserver/render"
	"github.com/jmoiron/sqlx"
	"net/http"
	"sync"
)

type dataContext struct {
	sync.RWMutex
	db  *sqlx.DB
	cfg *config.Source
}

func newDataContext(cfg *config.Source) *dataContext {
	return &dataContext{cfg: cfg}
}

func (c *dataContext) Get() (*sqlx.DB, error) {
	if db := c.tryGet(); db == nil {
		if err := c.tryInit(); err != nil {
			return nil, err
		} else {
			return c.Get() // init success, so retry getting
		}
	} else {
		return db, nil
	}
}
func (c *dataContext) tryGet() *sqlx.DB {
	c.RLock()
	defer c.RUnlock()

	return c.db
}
func (c *dataContext) tryInit() error {
	c.Lock()
	defer c.Unlock()

	if c.db != nil {
		return nil
	}

	if db, err := data.Connect(c.cfg); err != nil {
		return err
	} else {
		c.db = db
		return nil
	}
}

func AddDataContext(cfg *config.Source) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		dc := newDataContext(cfg)

		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if db, err := dc.Get(); err != nil {
				render.Error(resp, req, 500, err)
			} else {
				h.ServeHTTP(resp, req.WithContext(
					data.NewContext(req.Context(), db)))
			}
		})
	}
}
