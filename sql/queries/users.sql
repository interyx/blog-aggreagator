-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES(
  $1,
  $2,
  $3,
  $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE name = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users *;

-- name: GetAllUsers :many
SELECT * FROM users;

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
