package database

import (
	"database/sql"

	"github.com/aga-absolut/url-cutter/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBPostgreSQl struct {
	shortURL    string
	originalURL string
	config      *config.Config
}

func NewDBPostgreSQl(config *config.Config) *DBPostgreSQl {
	return &DBPostgreSQl{config: config}
}

func (s *DBPostgreSQl) Set(shortURL, originalURL string) error {
	db, err := sql.Open("pgx", s.config.DBDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS urls (
		short_url TEXT NOT NULL PRIMARY KEY,
		original_url TEXT NOT NULL
	);`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO urls VALUES ($1, $2)`, shortURL, originalURL)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBPostgreSQl) Get(shortURL string) (string, bool) {
	db, err := sql.Open("pgx", s.config.DBDSN)
	if err != nil {
		return "", false
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return "", false
	}

	row := db.QueryRow(`SELECT original_url FROM urls $1`, shortURL)

	var originalURL string
	err = row.Scan(&originalURL)
	if err != nil {
		return "", false
	}
	return originalURL, true
}
