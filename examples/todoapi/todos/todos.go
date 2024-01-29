package todos

import (
	"context"
	"encoding/json"
	"time"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/data"
	. "fx.prodigy9.co/examples/todoapi/gen/todoapi/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

var App = app.Build().
	Description("Basic username/password authentication").
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
	todo, sql := &Todo{}, Todos.
		SELECT(Todos.AllColumns).
		WHERE(
			Todos.UserID.EQ(postgres.Int64(userID)).
				AND(Todos.ID.EQ(postgres.Int64(id))),
		).
		LIMIT(1)

	err := data.GetSQL(ctx, todo, sql)
	return todo, err
}

func GetTodosByUserID(ctx context.Context, userID int64) ([]*Todo, error) {
	var todos []*Todo
	sql := Todos.
		SELECT(Todos.AllColumns).
		WHERE(Todos.UserID.EQ(postgres.Int64(userID))).
		ORDER_BY(Todos.ID.ASC())

	err := data.SelectSQL(ctx, &todos, sql)
	return todos, err
}

func DeleteTodo(ctx context.Context, userID, id int64) (*Todo, error) {
	todo, sql := &Todo{}, Todos.
		DELETE().
		WHERE(
			Todos.UserID.EQ(postgres.Int64(userID)).
				AND(Todos.ID.EQ(postgres.Int64(id))),
		).
		RETURNING(postgres.STAR)

	err := data.GetSQL(ctx, todo, sql)
	return todo, err
}
