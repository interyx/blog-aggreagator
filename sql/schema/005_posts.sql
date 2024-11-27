-- +goose Up
CREATE TABLE posts (
  id uuid PRIMARY KEY,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  title text NOT NULL,
  url text NOT NULL,
  UNIQUE(url),
  description text,
  published_at timestamp,
  feed_id uuid NOT NULL references feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
