package todos

import (
	"context"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/examples/todoapi/auth"
	. "fx.prodigy9.co/examples/todoapi/gen/todoapi/public/table"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/validate"
	"github.com/go-jet/jet/v2/postgres"
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

	return scope.GetSQL(out, Todos.
		INSERT(
			Todos.UserID,
			Todos.Title,
			Todos.Description).
		VALUES(
			user.ID,
			c.Title,
			c.Description,
		).
		RETURNING(postgres.STAR),
	)
}
