package database

import (
	"database/sql"
	"fmt"

	"github.com/aga-absolut/url-cutter/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectDB(config *config.Config) (*sql.DB, error) {
	ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, `postgres`, `absolute_1`, `mydb`)
	db, err := sql.Open(config.DbDSN, ps)
	if err != nil {
		return nil, err
	}
	return db, nil
}
