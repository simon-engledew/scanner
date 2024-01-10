package sqlh

import (
	"database/sql"
	"errors"
	"fmt"
)

// Rows allows wrappers for sql.Rows to be passed to the scanning functions.
type Rows interface {
	Row
	Close() error
	Next() bool
	Err() error
}

var _ Rows = &sql.Rows{}

// Pluck will scan the results of a query that produces a single column.
func Pluck[V any](rows Rows, queryErr error) (out []V, err error) {
	if queryErr != nil {
		return out, queryErr
	}
	return ScanV(rows, func(v *V, rows Row) error {
		return rows.Scan(v)
	})
}

// Iter calls fn for each successful call to rows.Next, closing the rows at the end.
func Iter(rows Rows, fn func() error) (err error) {
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	for rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}

		if err := fn(); err != nil {
			return fmt.Errorf("failed to scan rows: %w", err)
		}
	}

	return rows.Err()
}

// ScanV takes a function that can scan a given sql.Rows into []V.
func ScanV[P *V, V any](rows Rows, scan func(P, Row) error) (out []V, err error) {
	err = Iter(rows, func() error {
		var v V
		err := scan(&v, rows)
		if err == nil {
			out = append(out, v)
		}
		return err
	})
	return
}

// Scan takes a function that can scan a given sql.Rows into []P.
func Scan[P *V, V any](rows Rows, scan func(P, Row) error) (out []P, err error) {
	err = Iter(rows, func() error {
		var v V
		err := scan(&v, rows)
		if err == nil {
			out = append(out, &v)
		}
		return err
	})
	return
}
