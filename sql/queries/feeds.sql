-- name: CreateFeed :one
INSERT INTO feeds (id, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetAllFeeds :many
SELECT * FROM feeds;

-- name: CountFeeds :one
SELECT COUNT(*) FROM feeds;

-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;


-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = CURRENT_TIMESTAMP, 
updated_at = CURRENT_TIMESTAMP
WHERE id = $1;


-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;

