package todos

import (
	"context"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/examples/todoapi/auth"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/validate"
)

type CreateTodo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (c *CreateTodo) Validate() error {
	return validate.Multi(
		validate.Required("title", c.Title),
		validate.StrLen("title", c.Title, 3),
	)
}

func (c *CreateTodo) Execute(ctx context.Context, out any) (err error) {
	scope, cancel, err := data.NewScopeErr(ctx, &err)
	defer cancel()

	user := auth.UserFromContext(ctx)
	if user == nil {
		return httperrors.ErrUnauthorized
	}

	sql := `
		INSERT INTO todos (user_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING *`
	return scope.Get(out, sql, user.ID, c.Title, c.Description)
}
