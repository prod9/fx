package auth

import (
	"context"
	"time"

	"fx.prodigy9.co/data"
	. "fx.prodigy9.co/examples/todoapi/gen/todoapi/public/table"
	"fx.prodigy9.co/validate"
	"github.com/go-jet/jet/v2/postgres"
)

type CreateSession struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *CreateSession) Validate() error {
	return validate.Multi(
		validate.StrLen("username", c.Username, 3),
		validate.Required("username", c.Username),
		validate.Required("password", c.Password),
		validate.StrLen("password", c.Password, 8),
	)
}

func (c *CreateSession) Execute(ctx context.Context, out any) (err error) {
	scope, cancel, err := data.NewScopeErr(ctx, &err)
	defer cancel()

	user, err := GetUserByUsername(scope.Context(), c.Username)
	if err != nil {
		if data.IsNoRows(err) {
			return validate.NewFieldError("username", "not found", c.Username)
		} else {
			return err
		}
	}

	ok, err := user.ValidatePassword(c.Password)
	if err != nil {
		return err
	} else if !ok {
		return validate.NewFieldError("password", "invalid", nil)
	}

	token, err := GenerateSessionToken()
	if err != nil {
		return err
	}

	return scope.GetSQL(out, Sessions.
		INSERT(
			Sessions.UserID,
			Sessions.Token,
			Sessions.ExpiresAt).
		VALUES(
			user.ID,
			token,
			time.Now().Add(DefaultSessionAge),
		).
		RETURNING(postgres.STAR),
	)
}
