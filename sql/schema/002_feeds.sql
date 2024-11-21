-- +goose Up
CREATE TABLE feeds(
  id uuid primary key,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  name text,
  user_id uuid REFERENCES users (id) ON DELETE CASCADE,
  url text,
  UNIQUE(url)
);

-- +goose Down
DROP TABLE feeds;
