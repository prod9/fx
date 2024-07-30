package data

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type (
	Scope interface {
		Context() context.Context

		// End ends the scope. It is meant to be used inside a `defer` statement.
		End(*error)

		Get(dest interface{}, sql string, args ...interface{}) error
		Select(dest interface{}, sql string, args ...interface{}) error
		Exec(sql string, args ...interface{}) error
		Prepare(query string) (*sqlx.Stmt, error)

		GetSQL(interface{}, SQLGenerator) error
		SelectSQL(interface{}, SQLGenerator) error
		ExecSQL(SQLGenerator) error
		PrepareSQL(SQLGenerator) (*sqlx.Stmt, error)
	}

	txKey struct{}

	scopeImpl struct {
		ctx    context.Context
		cancel context.CancelFunc
		tx     *sqlx.Tx
		child  bool
	}
)

var _ Scope = scopeImpl{}

func getTx(ctx context.Context) (*sqlx.Tx, bool) {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx, true
	} else {
		return nil, false
	}
}

func setTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func newScope(ctx context.Context, db *sqlx.DB) (scope scopeImpl, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("tx: %w", err)
		}
	}()

	scope.ctx, scope.cancel = context.WithCancel(ctx)
	scope.tx, scope.child = getTx(ctx)
	if !scope.child {
		if scope.tx, err = db.BeginTxx(ctx, nil); err != nil {
			// not a child, first call, start new tx
			return scopeImpl{}, err
		}
	}

	scope.ctx = setTx(scope.ctx, scope.tx)
	return
}

func (s scopeImpl) Context() context.Context { return s.ctx }

func (s scopeImpl) End(err *error) {
	if !s.child {
		if *err == nil {
			*err = s.tx.Commit()
		} else {
			_ = s.tx.Rollback()
		}
	}
	if s.cancel != nil {
		s.cancel()
	}
}

func (s scopeImpl) Get(dest interface{}, sql string, args ...interface{}) error {
	return s.tx.GetContext(s.ctx, dest, sql, args...)
}
func (s scopeImpl) Select(dest interface{}, sql string, args ...interface{}) error {
	return s.tx.SelectContext(s.ctx, dest, sql, args...)
}
func (s scopeImpl) Exec(sql string, args ...interface{}) error {
	_, err := s.tx.ExecContext(s.ctx, sql, args...)
	return err
}
func (s scopeImpl) Prepare(query string) (*sqlx.Stmt, error) {
	return s.tx.Preparex(query)
}

func (s scopeImpl) GetSQL(out interface{}, sqlgen SQLGenerator) (err error) {
	sql, args := sqlgen.Sql()
	return s.Get(out, sql, args...)
}
func (s scopeImpl) SelectSQL(out interface{}, sqlgen SQLGenerator) (err error) {
	sql, args := sqlgen.Sql()
	return s.Select(out, sql, args...)
}
func (s scopeImpl) ExecSQL(sqlgen SQLGenerator) (err error) {
	sql, args := sqlgen.Sql()
	return s.Exec(sql, args...)
}
func (s scopeImpl) PrepareSQL(sqlgen SQLGenerator) (*sqlx.Stmt, error) {
	sql, _ := sqlgen.Sql()
	return s.Prepare(sql)
}
