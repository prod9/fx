package testutil

import (
	"strconv"
	"sync/atomic"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver"
	"fx.prodigy9.co/httpserver/controllers"
)

const StartingTestPort = 10000

var nextTestPort atomic.Int32

func ConfigureTest() *config.Source {
	return config.NewSource(
		&config.MemProvider{},
		config.DefaultSource().Vars(),
	)
}

func NextListenAddr() string {
	nextTestPort.CompareAndSwap(0, StartingTestPort)

	nextPort := nextTestPort.Add(1)
	return "0.0.0.0:" + strconv.FormatInt(int64(nextPort), 10)
}

func TestHTTPServer(cfg *config.Source) *httpserver.Server {
	addr := NextListenAddr()
	httpserver.New(cfg, 
}
