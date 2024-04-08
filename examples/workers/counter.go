package main

import (
	"context"
	"fx.prodigy9.co/data"
	"strconv"
)

type Counter struct {
	Name  string `db:"name"`
	Count int64  `db:"count"`
}

func GetAllCounters(ctx context.Context) ([]*Counter, error) {
	var counters []*Counter
	err := data.Select(ctx, &counters, `
		SELECT *
		FROM counters
		ORDER BY name ASC
		`)
	return counters, err
}

func GetCounterByName(ctx context.Context, name string) (*Counter, error) {
	counter := &Counter{}
	err := data.Get(ctx, counter, `
		SELECT *
		FROM counters
		WHERE name = $1
		`, name)
	return counter, err
}

func UpdateCounter(ctx context.Context, counter *Counter) error {
	err := data.Get(ctx, counter, `
		UPDATE counters
		SET count = $2
		WHERE name = $1
		RETURNING *
		`, counter.Name, counter.Count)
	return err
}

func CreateCounters(ctx context.Context, n int) (counters []*Counter, err error) {
	const SQL = `
	INSERT INTO counters (name, count)
	VALUES ($1, $2)
	RETURNING *`

	scope, cancel, err := data.NewScopeErr(ctx, &err)
	if err != nil {
		return nil, err
	}
	defer cancel()

	for i := 1; i <= n; i++ {
		counter := &Counter{}
		if err := scope.Get(counter, SQL, "counter"+strconv.Itoa(i), 0); err != nil {
			return nil, err
		}
		counters = append(counters, counter)
	}

	return counters, nil
}
