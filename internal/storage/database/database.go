package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

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
	config *config.Config
	db     *sql.DB
}

func NewDBPostgreSQL(config *config.Config) *DBPostgreSQL {
	db, err := sql.Open("pgx", config.DBDSN)
	if err != nil {
		log.Fatalf("cannot open db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			short_url TEXT NOT NULL PRIMARY KEY,
			original_url TEXT NOT NULL
		);
	`)
	if err != nil {
		log.Fatalf("error creating table: %v", err)
	}

	_, err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url
		ON urls(original_url);
	`)
	if err != nil {
		log.Fatalf("error creating index: %v", err)
	}

	return &DBPostgreSQL{
		db:     db,
		config: config,
	}
}

func (s *DBPostgreSQL) Set(shortURL, originalURL string) (string, error) {
	_, err := s.db.Exec(`INSERT INTO urls VALUES ($1, $2)`, shortURL, originalURL)
	if err != nil {
		var PgErr *pgconn.PgError
		if errors.As(err, &PgErr) {
			if PgErr.Code == pgerrcode.UniqueViolation {
				var shortKey string
				row := s.db.QueryRow(`SELECT short_url FROM urls WHERE original_url = $1`, originalURL)
				row.Scan(&shortKey)
				return shortKey, os.ErrExist
			}
		}
		return "", err
	}

	return "", nil
}

func (s *DBPostgreSQL) SetBatchURL(batch []ShotenBatchRequest) ([]ShortenResponseItem, error) {
	var responseItem ShortenResponseItem
	var response []ShortenResponseItem
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error add tx: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO urls VALUES ($1, $2)`)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error add stmt: %w", err)
	}
	defer stmt.Close()

	for _, v := range batch {
		shortKey := s.config.Generate()
		_, err = stmt.Exec(shortKey, v.OriginalURL)
		if err != nil {
			tx.Rollback()
			var PgErr *pgconn.PgError
			if errors.As(err, &PgErr) {
				if PgErr.Code == pgerrcode.UniqueViolation {
					return nil, fmt.Errorf("not unique URL: %w", err)
				}
			}
			return nil, fmt.Errorf("error request: %w", err)
		}

		responseItem.ShortURL = s.config.Host + "/" + shortKey
		responseItem.CorrelationID = v.CorrelationID
		response = append(response, responseItem)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error commit tx: %w", err)
	}

	return response, nil
}

func (s *DBPostgreSQL) Get(shortURL string) (string, bool) {
	row := s.db.QueryRow(`SELECT original_url FROM urls WHERE short_url = $1`, shortURL)

	var originalURL string
	err := row.Scan(&originalURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	return originalURL, true
}
