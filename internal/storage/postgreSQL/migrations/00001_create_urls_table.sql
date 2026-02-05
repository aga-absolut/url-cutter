-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
    short_url TEXT NOT NULL PRIMARY KEY,
    original_url TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_original_url;
DROP TABLE urls;
-- +goose StatementEnd