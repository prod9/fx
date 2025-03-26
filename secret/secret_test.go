package secret

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"

	"fx.prodigy9.co/config"
	"github.com/stretchr/testify/require"
)

const (
	TestSecret = "the cake is a lie"
	TestPhrase = "the enrichment center promises cake"
)

func TestHideReveal(t *testing.T) {
	cfg := config.NewSource(
		&config.MemProvider{},
		config.DefaultSource().Vars(),
	)

	config.Set(cfg, SecretConfig, TestSecret)

	secret, err := Hide(cfg, TestPhrase)
	require.NoError(t, err)
	require.NotEmpty(t, secret)
	require.NotEqual(t, TestPhrase, secret)

	phrase, err := Reveal(cfg, secret)
	require.NoError(t, err)
	require.NotEmpty(t, phrase)
	require.Len(t, phrase, len(TestPhrase))
	require.Equal(t, TestPhrase, phrase)
}

func TestHideReveal_LongString(t *testing.T) {
	cfg := config.NewSource(
		&config.MemProvider{},
		config.DefaultSource().Vars(),
	)

	config.Set(cfg, SecretConfig, TestSecret)

	buf := &bytes.Buffer{}
	_, err := io.CopyN(buf, rand.Reader, 1<<10) // 1KB
	require.NoError(t, err)

	plaintext := hex.EncodeToString(buf.Bytes())

	secret, err := Hide(cfg, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, secret)
	require.NotEqual(t, plaintext, secret)

	output, err := Reveal(cfg, secret)
	require.NoError(t, err)
	require.NotEmpty(t, output)
	require.Len(t, output, len(plaintext))
	require.Equal(t, plaintext, output)
}

func TestHideReveal_ChangedSecret(t *testing.T) {
	cfg := config.NewSource(
		&config.MemProvider{},
		config.DefaultSource().Vars(),
	)

	secret, err := Hide(cfg, TestPhrase)
	require.NoError(t, err)
	require.NotEmpty(t, secret)
	require.NotEqual(t, TestPhrase, secret)

	_container = container{}
	config.Set(cfg, SecretConfig, TestSecret[:len(TestSecret)>>1])

	_, err = Reveal(cfg, secret)
	require.NotNil(t, err)
}
