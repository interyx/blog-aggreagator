-- +goose Up
CREATE TABLE feeds(
  id uuid primary key,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  name text NOT NULL,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  url text NOT NULL,
  UNIQUE(url)
);

-- +goose Down
DROP TABLE feeds;
