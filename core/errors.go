package core

import (
	"database/sql"
	"github.com/lib/pq"
)

const (
	SuccessCompleted = iota + 100
	SystemError
	ForeignKeyError
)

type WrapError struct {
	Err  error
	Code int
}

func (w *WrapError) Error() string {
	return w.Err.Error()
}

func ErrNoRows(err error) bool {
	return err == sql.ErrNoRows
}

func ErrForeignKey(err error) bool {
	if pgErr, isPgErr := err.(*pq.Error); isPgErr {
		return pgErr.Code == "23503"
	} else {
		return false
	}
}
