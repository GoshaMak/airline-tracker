package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	driver "github.com/go-sql-driver/mysql"
	"github.com/samber/do/v2"
)

func NewMySQLDB(i do.Injector) (*sql.DB, error) {
	return newDB(context.Background())
}

func newDB(ctx context.Context) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_NAME"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to open mysql connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to ping mysql: %w", err)
	}

	return db, nil
}

func CloseConnection(db *sql.DB) error {
	return db.Close()
}

func IsDuplicate(err error) bool {
	var mysqlErr *driver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func IsNotFoundSignal(err error, message string) bool {
	var mysqlErr *driver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Message == message
}
