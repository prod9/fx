package auth

import (
	"context"
	"time"

	"fx.prodigy9.co/data"
	. "fx.prodigy9.co/examples/todoapi/gen/todoapi/public/table"
	"fx.prodigy9.co/validate"
	"github.com/go-jet/jet/v2/postgres"
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
	return data.GetSQL(ctx, out, Sessions.
		UPDATE(Sessions.ExpiresAt).
		SET(Sessions.ExpiresAt, time.Now()).
		WHERE(Sessions.Token.EQ(postgres.String(c.Token))).
		RETURNING(postgres.STAR),
	)
}
