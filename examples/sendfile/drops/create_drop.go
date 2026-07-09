package drops

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"fx.prodigy9.co/data"
)

type CreateDrop struct{}

func (c *CreateDrop) Execute(ctx context.Context, out any) error {
	token, err := randomToken()
	if err != nil {
		return err
	}

	sql := `INSERT INTO drops (token) VALUES ($1) RETURNING *`
	return data.Get(ctx, out, sql, token)
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
