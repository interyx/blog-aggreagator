-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, user_id, url)
VALUES(
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
  )
  RETURNING *;

-- name: GetAllFeeds :many
SELECT feeds.name, feeds.url, users.name FROM feeds
INNER JOIN users ON user_id = users.id;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;
