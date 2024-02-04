package page

import (
	"context"
	"errors"
	"strconv"

	"fx.prodigy9.co/data"
)

const DefaultPageSize = 20

var ErrNilOut = errors.New("page: out must be a pointer to a Page")

type Page[T any] struct {
	Meta
	Data []T `json:"data"`

	TotalPages int `json:"total_pages"`
	TotalRows  int `json:"total_rows"`
}

func Select[T any](ctx context.Context, out *Page[T], meta Meta, sql string, args ...any) (err error) {
	if out == nil {
		return ErrNilOut
	}
	if meta.RowsPerPage <= 0 {
		meta.RowsPerPage = DefaultPageSize
	}
	if meta.Page <= 0 {
		meta.Page = 1
	}

	scope, cancel, err := data.NewScopeErr(ctx, &err)
	if err != nil {
		return err
	}
	defer cancel()

	var (
		limitArgIdx  = len(args) + 1
		offsetArgidx = len(args) + 2

		prefix   = "WITH source AS (" + sql + ")\n"
		countSQL = prefix + "SELECT COUNT(*) FROM source"
		dataSQL  = prefix + "SELECT * FROM source" +
			" LIMIT $" + strconv.Itoa(limitArgIdx) +
			" OFFSET $" + strconv.Itoa(offsetArgidx)

		count    int
		offset   = (meta.Page - 1) * meta.RowsPerPage
		dataArgs = append(args, meta.RowsPerPage, offset)
	)

	if err := scope.Get(&count, countSQL, args...); err != nil {
		return err
	}
	if count == 0 {
		*out = Page[T]{
			Meta:       meta,
			Data:       []T{},
			TotalPages: 0,
			TotalRows:  0,
		}
		return nil
	}

	var data []T
	if err := scope.Select(&data, dataSQL, dataArgs...); err != nil {
		return err
	}

	*out = Page[T]{
		Meta:       meta,
		Data:       data,
		TotalPages: (count + meta.RowsPerPage - 1) / meta.RowsPerPage,
		TotalRows:  count,
	}
	return nil
}

func SelectSQL[T any](ctx context.Context, out *Page[T], meta Meta, sqlgen data.SQLGenerator) (err error) {
	sql, args := sqlgen.Sql()
	return Select(ctx, out, meta, sql, args...)
}
