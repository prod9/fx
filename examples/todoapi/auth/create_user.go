package auth

import (
	"context"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/passwords"
	"fx.prodigy9.co/validate"
)

type CreateUser struct {
	Username             string `json:"username"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func (c *CreateUser) Validate() error {
	return validate.Multi(
		validate.StrLen("username", c.Username, 3),
		validate.Required("username", c.Username),
		validate.Required("password", c.Password),
		validate.StrLen("password", c.Password, 8),
		validate.Required("password_confirmation", c.PasswordConfirmation),
		validate.FieldsMatch("password", c.Password, "password_confirmation", c.PasswordConfirmation),
	)
}

func (c *CreateUser) Execute(ctx context.Context, out any) (err error) {
	scope, cancel, err := data.NewScopeErr(ctx, &err)
	defer cancel()

	var pwdhash string
	pwdhash, err = passwords.Hash(c.Password)

	sql := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING *`
	return scope.Get(out, sql, c.Username, pwdhash)
}
