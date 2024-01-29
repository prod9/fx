package todos

import (
	"context"
	"errors"
	"time"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/examples/todoapi/auth"
	. "fx.prodigy9.co/examples/todoapi/gen/todoapi/public/table"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/validate"
	"github.com/go-jet/jet/v2/postgres"
)

type UpdateTodo struct {
	ID        int64 `json:"-"`
	Completed bool  `json:"completed"`
}

func (u *UpdateTodo) Validate() error {
	return validate.Positive("id", u.ID)
}

func (u *UpdateTodo) Execute(ctx context.Context, out any) (err error) {
	scope, cancel, err := data.NewScopeErr(ctx, &err)
	defer cancel()

	user := auth.UserFromContext(ctx)
	if user == nil {
		return httperrors.ErrUnauthorized
	}

	todo, err := GetTodo(scope.Context(), user.ID, u.ID)
	if err != nil {
		return err
	}

	if todo.IsCompleted() == u.Completed {
		// nothing to do, already completed, just copy the output
		if out != nil {
			if outtodo, ok := out.(*Todo); ok && outtodo != nil {
				*outtodo = *todo
				return nil
			} else if outtodo != nil {
				return errors.New("invalid output type")
			} else {
				// no output
				return nil
			}
		}
	}

	// otherwise, update todo to be completed (or not)
	var sql data.SQLGenerator
	if u.Completed {
		sql = Todos.
			UPDATE(Todos.CompletedAt).
			SET(Todos.CompletedAt.SET(postgres.TimestampzT(time.Now()))).
			WHERE(
				Todos.UserID.EQ(postgres.Int64(user.ID)).
					AND(Todos.ID.EQ(postgres.Int64(u.ID))),
			).
			RETURNING(postgres.STAR)
	} else {
		sql = Todos.
			UPDATE(Todos.CompletedAt).
			SET(Todos.CompletedAt.SET(nil)).
			WHERE(
				Todos.UserID.EQ(postgres.Int64(user.ID)).
					AND(Todos.ID.EQ(postgres.Int64(u.ID))),
			).
			RETURNING(postgres.STAR)
	}

	return scope.GetSQL(out, sql)
}
