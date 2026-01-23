package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type ShotenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
type DBPostgreSQL struct {
	ShortURL    string
	OriginalURL string
	config      *config.Config
}

func ConnectDBPostgreSQL(config *config.Config) *sql.DB {
	if db, err := sql.Open("pgx", config.DBDSN); err == nil {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
		short_url TEXT NOT NULL PRIMARY KEY,
		original_url TEXT NOT NULL);`)
		if err != nil {
			log.Print("Error creating table")
			return nil
		}
		_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls(original_url)`)
        if err != nil {
            log.Printf("Error creating unique index: %v", err)
        }
		return db
	}
	return nil
}

func NewDBPostgreSQL(config *config.Config) *DBPostgreSQL {
	return &DBPostgreSQL{config: config}
}

func (s *DBPostgreSQL) Set(shortURL, originalURL string) error {
	db, err := sql.Open("pgx", s.config.DBDSN)
	if err != nil {
		return err
	}
	defer db.Close()

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
		var PgErr *pgconn.PgError
		if errors.As(err, &PgErr) {
			if PgErr.Code == pgerrcode.UniqueViolation {
				return fmt.Errorf("not unique URL: %v", err)
			}
		}
		return err
	}

	return nil
}

func (s *DBPostgreSQL) Get(shortURL string) (string, bool) {
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
