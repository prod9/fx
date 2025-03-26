package middlewares

import (
	"testing"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/internal/testutil"
	"github.com/rs/cors"
)

func TestCORS(t *testing.T) {
	cfg := config.NewSource(
		&config.MemProvider{},
		config.DefaultSource().Vars(),
	)

	config.Set(cfg, httpserver.ListenAddrConfig, testutil.NextListenAddr())

	fragment := httpserver.NewFragment(nil, nil).
		AddMiddlewares(CORS(cors.Options{AllowedHeaders: []string{"PUT", "DELETE"}})).
		AddControllers(controllers.Home{})

	server := httpserver.NewWithFragments(cfg, []*httpserver.Fragment{fragment})
}
