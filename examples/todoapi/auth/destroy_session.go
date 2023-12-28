package auth

import (
	"context"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/validate"
)

type DestroySession struct {
	Token string `json:"token"`
}

func (c *DestroySession) Validate() error {
	return validate.Multi(
		validate.Required("token", c.Token),
		validate.StrLen("token", c.Token, SessionTokenBytes), // actually its bytes*(3/2) since it's base64
	)
}

func (c *DestroySession) Execute(ctx context.Context, out any) (err error) {
	sql := `
		UPDATE sessions
		SET expires_at = CURRENT_TIMESTAMP
		WHERE token = $1
		RETURNING *`
	return data.Get(ctx, out, sql, c.Token)
}
