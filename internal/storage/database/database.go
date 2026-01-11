package database

import (
	"database/sql"

	"github.com/aga-absolut/url-cutter/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectDB(config *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.DBDSN)
	if err != nil {
		return nil, err
	}
	return db, nil
}
