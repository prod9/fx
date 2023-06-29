package data

import (
	"fmt"
	"fx.prodigy9.co/structs"
	"strings"
)

type ListData[T any] struct {
	List  []T  `json:"list"`
	Total *int `json:"total"`
}

func NewListResponce[T any]() *ListData[T] {
	return &ListData[T]{
		List:  []T{},
		Total: nil,
	}
}

type QueryBuilder struct {
	Sql    string
	Filter structs.ParsedStruct
	Args   []interface{}
}

func (qb *QueryBuilder) SelectColumns(columns string) *QueryBuilder {
	qb.Sql = strings.Replace(qb.Sql, "{columns}", columns, 1)
	return qb
}

func (qb *QueryBuilder) Count() *QueryBuilder {
	qb.Sql = strings.Replace(qb.Sql, "{columns}", "COUNT(id)", 1)
	qb.Sql = strings.Replace(qb.Sql, "{paginate}", "", 1)
	qb.Sql = strings.Replace(qb.Sql, "{order}", "", 1)
	return qb.Where()
}

func (qb *QueryBuilder) Update() *QueryBuilder {
	return qb.withClause("{update}", "column", ", ", "")
}

func (qb *QueryBuilder) QueryParams() (string, []interface{}) {
	return qb.Sql, qb.Args
}

func (qb *QueryBuilder) Paginate() *QueryBuilder {
	page, size := 0, 0
	if tag := qb.Filter.FindFieldByTag("fx", "page", nil); tag != nil {
		page, _ = tag.Value.(int)
	}
	if tag := qb.Filter.FindFieldByTag("fx", "size", nil); tag != nil {
		size, _ = tag.Value.(int)
	}
	paginate := ""
	if page > 0 && size > 0 {
		paginate = fmt.Sprintf(" OFFSET $%v LIMIT $%v", len(qb.Args)+1, len(qb.Args)+2)
		qb.Args = append(qb.Args, (page-1)*size, size)
	}
	qb.Sql = strings.Replace(qb.Sql, "{paginate}", paginate, 1)

	return qb
}

func (qb *QueryBuilder) Where() *QueryBuilder {
	return qb.withClause("{where}", "where", " AND ", "WHERE")
}

func (qb *QueryBuilder) withClause(replace string, tagName string, separator string, prefix string) *QueryBuilder {
	parts := []string{}
	for _, field := range qb.Filter.Fields {
		tag := field.GetTag("fx", tagName)
		if field.IsZero || tag == nil {
			continue
		}
		column := tag.Value
		if column == "" {
			column = field.Name
		}
		condition := fmt.Sprintf("%s = $%v", strings.ToLower(column), len(qb.Args)+1)
		parts = append(parts, condition)
		qb.Args = append(qb.Args, field.Value)
	}
	clause := ""
	if len(parts) > 0 {
		clause = prefix + " " + strings.Join(parts, separator)
	}

	qb.Sql = strings.Replace(qb.Sql, replace, clause, 1)

	return qb
}

func (qb *QueryBuilder) Order() *QueryBuilder {
	order := ""
	if tag := qb.Filter.FindFieldByTag("fx", "order", nil); tag != nil {
		order = fmt.Sprintf(" ORDER BY id %s", tag.Value)
	}
	qb.Sql = strings.Replace(qb.Sql, "{order}", order, 1)
	return qb
}

type ResourceInterface interface {
	GetTableName() string
}
