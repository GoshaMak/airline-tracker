package postgres

import "errors"

type PgErrorCode string

const (
	RecordAlreadyExistsErrCode PgErrorCode = "23505"
)

var (
	ErrInsertFailure       = errors.New("Can't insert")
	ErrRecordAlreadyExists = errors.New("Record already exists")
	ErrRecordNotExists     = errors.New("Record doesn't exist")
)
