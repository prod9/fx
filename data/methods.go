package data

import "context"

// small shim to support sql generators like go-jet
type SQLGenerator interface {
	Sql() (string, []any)
}

func Get(ctx context.Context, out any, sql string, args ...any) (err error) {
	return Run(ctx, func(s Scope) error { return s.Get(out, sql, args...) })
}
func Select(ctx context.Context, out any, sql string, args ...any) (err error) {
	return Run(ctx, func(s Scope) error { return s.Select(out, sql, args...) })
}
func Exec(ctx context.Context, sql string, args ...any) error {
	return Run(ctx, func(s Scope) error { return s.Exec(sql, args...) })
}

func GetSQL(ctx context.Context, out any, sqlgen SQLGenerator) (err error) {
	sql, args := sqlgen.Sql()
	return Get(ctx, out, sql, args...)
}
func SelectSQL(ctx context.Context, out any, sqlgen SQLGenerator) (err error) {
	sql, args := sqlgen.Sql()
	return Select(ctx, out, sql, args...)
}
func ExecSQL(ctx context.Context, sqlgen SQLGenerator) (err error) {
	sql, args := sqlgen.Sql()
	return Exec(ctx, sql, args...)
}

func Run(ctx context.Context, action func(s Scope) error) (err error) {
	var scope Scope
	if scope, err = NewScope(ctx, nil); err != nil {
		return
	} else {
		defer scope.End(&err)
		return action(scope)
	}
}
