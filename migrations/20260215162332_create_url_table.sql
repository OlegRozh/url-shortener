-- +goose Up
CREATE TABLE IF NOT EXISTS url (
                                   id SERIAL PRIMARY KEY,
                                   alias TEXT NOT NULL UNIQUE,
                                   url TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
CREATE INDEX IF NOT EXISTS idx_url ON url(url);

-- +goose Down
DROP TABLE IF EXISTS url;
