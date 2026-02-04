package database

import (
	"context"
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
			user_id INTEGER NOT NULL,
			is_deleted BOOLEAN DEFAULT FALSE
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

func (s *DBPostgreSQL) Set(ctx context.Context, shortURL, originalURL string, userID int) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", fmt.Errorf("error add tx: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO urls (short_url, original_url, user_id)
    VALUES ($1, $2, $3)`, shortURL, originalURL, userID)
	if err != nil {
		var PgErr *pgconn.PgError
		if errors.As(err, &PgErr) {
			if PgErr.Code == pgerrcode.UniqueViolation {
				var shortKey string
				row := tx.QueryRowContext(ctx, `SELECT short_url FROM urls WHERE original_url = $1`, originalURL)
				if err := row.Scan(&shortKey); err != nil {
					return "", fmt.Errorf("error scaning query row: %w", err)
				}
				if err := tx.Commit(); err != nil {
					return "", fmt.Errorf("error commit query request: %w", err)
				}
				return shortKey, os.ErrExist
			}
		}
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("error commit insert request: %w", err)
	}
	return "", nil
}

func (s *DBPostgreSQL) SetBatchURL(ctx context.Context, batch []model.ShotenBatchRequest, userID int) ([]model.ShortenResponseItem, error) {
	var response []model.ShortenResponseItem
	stmt, err := s.db.Prepare(`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`)
	if err != nil {
		return nil, fmt.Errorf("error add stmt: %w", err)
	}
	defer stmt.Close()

	for _, v := range batch {
		shortKey := config.Generate(v.OriginalURL)
		_, err = stmt.ExecContext(ctx, shortKey, v.OriginalURL, userID)
		if err != nil {
			var PgErr *pgconn.PgError
			if errors.As(err, &PgErr) {
				if PgErr.Code == pgerrcode.UniqueViolation {
					return nil, fmt.Errorf("not unique URL: %w", err)
				}
			}
			return nil, fmt.Errorf("error request: %w", err)
		}

		response = append(response, model.ShortenResponseItem{
			ShortURL:      s.config.Host + "/" + shortKey,
			CorrelationID: v.CorrelationID,
		})
	}
	
	return response, nil
}

func (s *DBPostgreSQL) Get(ctx context.Context, shortURL string) (string, bool) {
	var (
		deletedFlag bool
		originalURL string
	)

	row := s.db.QueryRowContext(ctx, `SELECT is_deleted, original_url FROM urls WHERE short_url = $1`, shortURL)

	err := row.Scan(&deletedFlag, &originalURL)
	if deletedFlag {
		return "", false
	}

	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	return originalURL, true
}

func (s *DBPostgreSQL) GetByUserID(ctx context.Context, userID int) (map[string]string, error) {
	var shortURL string
	var originalURL string
	mapURLs := make(map[string]string)

	rows, err := s.db.QueryContext(ctx, `SELECT short_url, original_url FROM urls WHERE user_id = $1`, userID)
	if err != nil {
		s.logger.Errorw("request completion error", "error", err, "userID", userID)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&shortURL, &originalURL)
		if err != nil {
			s.logger.Errorw("String scaning error", "error", err)
			return nil, err
		}
		mapURLs[shortURL] = originalURL
	}

	if err := rows.Err(); err != nil {
		s.logger.Errorw("error rows", "error", err)
		return nil, err
	}

	return mapURLs, nil
}

func (s *DBPostgreSQL) DeletedFlag(ctx context.Context, shortURL string) error {
	query := `UPDATE urls SET is_deleted = true WHERE short_url = $1`
	_, err := s.db.ExecContext(ctx, query, shortURL)
	return err
}
