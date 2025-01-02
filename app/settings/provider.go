package settings

import (
	"context"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"github.com/jmoiron/sqlx"
)

type Provider struct {
	db *sqlx.DB
}

var _ config.Provider = &Provider{}

func NewProvider(ctx context.Context) (*Provider, error) {
	cfg := config.FromContext(ctx)
	if cfg == nil {
		cfg = config.Configure()
	}

	db := data.FromContext(ctx)
	if db == nil {
		if d, err := data.Connect(cfg); err != nil {
			return nil, err
		} else {
			db = d
		}
	}

	return &Provider{db}, nil
}

func (p *Provider) Initialize() error {
	return p.db.Ping()
}

func (p *Provider) Get(name string) (string, bool, error) {
	settings, err := Get(p.dbContext(), name)
	switch {
	case data.IsNoRows(err):
		return "", false, nil
	case err != nil:
		return "", false, err
	default:
		return settings.Value, true, nil
	}
}

func (p *Provider) Set(name string, val string) error {
	_, err := Set(p.dbContext(), name, val)
	return err
}

func (p *Provider) dbContext() context.Context {
	// TODO: Maybe save in struct?
	ctx := context.Background()
	return data.NewContext(ctx, p.db)
}
