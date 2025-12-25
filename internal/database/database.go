package database

import (
	"database/sql"
)

func Connect(dbUrl string) (*sql.DB, error) {
	return sql.Open("postgres", dbUrl)
}
