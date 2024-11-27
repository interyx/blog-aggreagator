-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES(
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8
  )
  RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.title, posts.description FROM posts
WHERE posts.feed_id = (
  SELECT feed_follows.feed_id FROM feed_follows 
  INNER JOIN users ON user_id = users.id
  INNER JOIN feeds ON feed_id = feeds.id
  WHERE users.name = $1
)
ORDER BY published_at DESC
LIMIT $2;

