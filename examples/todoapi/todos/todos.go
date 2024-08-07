package todos

import (
	"context"
	"encoding/json"
	"time"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/data"
	pager "fx.prodigy9.co/data/page"
	"fx.prodigy9.co/examples/todoapi/auth"
)

var App = app.Build().
	Description("Basic username/password authentication").
	Middlewares(auth.RequireSession).
	Controllers(Ctr{})

type Todo struct {
	ID     int64 `json:"id" db:"id"`
	UserID int64 `json:"user_id" db:"user_id"`

	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

func (t *Todo) IsCompleted() bool {
	return t.CompletedAt != nil &&
		!t.CompletedAt.IsZero()
}

func (t *Todo) MarshalJSON() ([]byte, error) {
	shim := struct {
		Todo
		Completed bool `json:"completed"`
	}{
		Todo:      *t,
		Completed: t.IsCompleted(),
	}

	return json.Marshal(shim)
}

func GetTodo(ctx context.Context, userID, id int64) (*Todo, error) {
	todo := &Todo{}
	sql := `
		SELECT *
		FROM todos
		WHERE user_id = $1 AND id = $2
		LIMIT 1`
	err := data.Get(ctx, todo, sql, userID, id)
	return todo, err
}

func GetTodosByUserID(ctx context.Context, page pager.Meta, userID int64) (*pager.Page[*Todo], error) {
	sql := `
		SELECT *
		FROM todos
		WHERE user_id = $1
		ORDER BY id ASC`

	out := &pager.Page[*Todo]{}
	err := pager.Select(ctx, out, page, sql, userID)
	return out, err
}

func DeleteTodo(ctx context.Context, userID, id int64) (*Todo, error) {
	todo := &Todo{}
	sql := `
		DELETE FROM todos
		WHERE user_id = $1 AND id = $2
		RETURNING *`
	err := data.Get(ctx, todo, sql, userID, id)
	return todo, err
}
