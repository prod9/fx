package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"fx.prodigy9.co/config"
)

var (
	SecretConfig = config.Str("SECRET")

	_container container
	_          = fmt.Sprint // debug
)

func Hide(cfg *config.Source, data string) (string, error) {
	gcm, err := setupAESGCM(cfg)
	if err != nil {
		return "", err
	}

	var (
		nonceLen = gcm.NonceSize()
		inbuf    = []byte(data)
		outbuf   = make([]byte, nonceLen+gcm.Overhead()+len(inbuf))
	)
	if n, err := rand.Read(outbuf[:nonceLen]); err != nil {
		return "", err
	} else if n != nonceLen {
		return "", errors.New("nonce generation error")
	}

	gcm.Seal(outbuf[nonceLen:nonceLen], outbuf[:nonceLen], inbuf, nil)
	return base64.URLEncoding.EncodeToString(outbuf), nil
}

func Reveal(cfg *config.Source, data string) (string, error) {
	gcm, err := setupAESGCM(cfg)
	if err != nil {
		return "", err
	}

	buf, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	var (
		nonceLen = gcm.NonceSize()
		nonce    = buf[:nonceLen]
		inbuf    = buf[nonceLen:]
		outbuf   = make([]byte, 0, len(buf)-gcm.Overhead()-nonceLen)
	)

	outbuf, err = gcm.Open(outbuf, nonce, inbuf, nil)
	if err != nil {
		return "", err
	} else {
		return string(outbuf), nil
	}
}

func prepareContainer(cfg *config.Source) error {
	if !_container.shouldInit() {
		return _container.lastErr()
	}

	secret := config.Get(cfg, SecretConfig)
	return _container.ensureInit([]byte(secret))
}

func setupAESGCM(cfg *config.Source) (cipher.AEAD, error) {
	if err := prepareContainer(cfg); err != nil {
		return nil, err
	}

	key := _container.getKey()
	if block, err := aes.NewCipher(key[:]); err != nil {
		return nil, err
	} else {
		return cipher.NewGCM(block)
	}
}
