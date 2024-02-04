package page

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Meta struct {
	Page        int `json:"page"`
	RowsPerPage int `json:"rows_per_page"`
}

func FromRequest(req *http.Request) Meta {
	return FromQuery(req.URL.Query())
}

func FromQuery(query url.Values) Meta {
	var (
		err        error
		rawPage    string
		rawPerPage string

		page    int
		perPage int
	)

	if query.Has("page") {
		rawPage = strings.TrimSpace(query.Get("page"))
	}
	if query.Has("per_page") {
		rawPerPage = strings.TrimSpace(query.Get("per_page"))
	}

	if rawPage == "" {
		page = 1
	} else if page, err = strconv.Atoi(rawPage); err != nil {
		page = 1
	}

	if rawPerPage == "" {
		perPage = DefaultPageSize
	} else if perPage, err = strconv.Atoi(rawPerPage); err != nil {
		perPage = DefaultPageSize
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = DefaultPageSize
	}

	return Meta{
		Page:        page,
		RowsPerPage: perPage,
	}
}
