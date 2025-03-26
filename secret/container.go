package secret

import (
	"crypto/sha256"
	"errors"
	"sync"

	"golang.org/x/crypto/hkdf"
)

type container struct {
	sync.RWMutex
	initialized bool

	err  error
	salt []byte
	key  [32]byte
}

func (c *container) ensureInit(secret []byte) error {
	if !c.tryInit(secret) {
		return c.lastErr()
	} else { // tryInit succeeded
		return nil
	}
}

func (c *container) shouldInit() bool {
	c.RLock()
	defer c.RUnlock()
	return !c.initialized
}
func (c *container) keyLen() int {
	return len(c.key)
}
func (c *container) getKey() [32]byte {
	c.RLock()
	defer c.RUnlock()
	return c.key
}
func (c *container) getSalt() []byte {
	c.RLock()
	defer c.RUnlock()
	return c.salt
}
func (c *container) lastErr() error {
	c.RLock()
	defer c.RUnlock()
	return c.err
}

func (c *container) tryInit(secret []byte) bool {
	c.Lock()
	defer c.Unlock()

	if c.initialized {
		return c.err == nil
	}

	// could not figure out a simple way to share salt between processes so just going to
	// use fixed salts for now
	c.salt = []byte("thespicemustflow")

	r := hkdf.New(sha256.New, secret, c.salt, nil)
	n, err := r.Read(c.key[:])
	if err != nil {
		c.err = err
		c.initialized = true
		return false
	} else if n != len(c.key) {
		c.err = errors.New("generated key is too short")
		c.initialized = true
		return false
	}

	return true
}
