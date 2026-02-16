package postgres

type PgErrorCode string

const (
	RecordAlreadyExistsErrCode PgErrorCode = "23505"
)
