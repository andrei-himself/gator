-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id, last_fetched_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    NULL
)
RETURNING *;

-- name: DeleteFeeds :exec
DELETE FROM feeds;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE URL = $1; 

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $2,
    updated_at = $2
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at NULLS FIRST
FETCH FIRST 1 ROW ONLY;