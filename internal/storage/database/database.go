package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type DBPostgreSQL struct {
	config *config.Config
	db     *sql.DB
	logger zap.SugaredLogger
}

func NewDBPostgreSQL(config *config.Config, logger zap.SugaredLogger) *DBPostgreSQL {
	db, err := sql.Open("pgx", config.DBDSN)
	if err != nil {
		log.Fatalf("cannot open db: %v", err)
	}

	// db.Exec("DROP TABLE IF EXISTS urls CASCADE;")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			short_url TEXT NOT NULL PRIMARY KEY,
			original_url TEXT NOT NULL,
			user_id INTEGER NOT NULL
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
		logger: logger,
	}
}

func (s *DBPostgreSQL) Set(shortURL, originalURL string) (string, error) {
	_, err := s.db.Exec(`INSERT INTO urls (short_url, original_url, user_id)
    VALUES ($1, $2, $3)`, shortURL, originalURL, config.UserID)
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

func (s *DBPostgreSQL) SetBatchURL(batch []model.ShotenBatchRequest) ([]model.ShortenResponseItem, error) {
	var responseItem model.ShortenResponseItem
	var response []model.ShortenResponseItem
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error add tx: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error add stmt: %w", err)
	}
	defer stmt.Close()

	for _, v := range batch {
		shortKey := s.config.Generate()
		_, err = stmt.Exec(shortKey, v.OriginalURL, config.UserID)
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

func (s *DBPostgreSQL) GetByUserID(userID int) (map[string]string, error) {
	var shortURL string
	var originalURL string
	mapURLs := make(map[string]string)

	rows, err := s.db.Query(`SELECT short_url, original_url FROM urls WHERE user_id = $1`, userID)
	if err != nil {
		s.logger.Errorw("request completion error", "error", err, "userID", userID)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&shortURL, &originalURL)
		if err != nil {
			s.logger.Errorw("String scaning error", "error", err)
		}
		mapURLs[shortURL] = originalURL
	}

	if err := rows.Err(); err != nil{
		s.logger.Errorw("error rows", "error", err)
		return nil, err
	}
	
	return mapURLs, nil
}
