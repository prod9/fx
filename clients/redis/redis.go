// Package redis hands out shared go-redis clients, one per distinct URL — the single
// place FX dials Redis. A *goredis.Client is itself a connection pool and dials lazily
// on first command, so acquisition is cheap and consumers (cache, pubsub) share one pool
// per URL automatically. Clients are owned by this package: consumers must never Close
// one.
package redis

import (
	"sync"

	goredis "github.com/redis/go-redis/v9"
)

var (
	mutex   sync.Mutex
	clients = map[string]*goredis.Client{}
)

// Client returns the shared client for url, creating it on first use.
func Client(url string) (*goredis.Client, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if client, ok := clients[url]; ok {
		return client, nil
	}

	opts, err := goredis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := goredis.NewClient(opts)
	clients[url] = client
	return client, nil
}
