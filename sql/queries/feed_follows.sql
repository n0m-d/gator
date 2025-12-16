-- name: CreateFeedFollow :one
WITH inserted AS (
    INSERT INTO feed_follows (id, user_id, feed_id, created_at, updated_at)
    VALUES ($1, $2, $3, NOW(), NOW())
    RETURNING id, user_id, feed_id, created_at, updated_at
)
SELECT 
    inserted.id,
    inserted.user_id,
    inserted.feed_id,
    inserted.created_at,
    inserted.updated_at,
    u.name AS user_name,
    fd.name AS feed_name
FROM inserted
JOIN users u ON u.id = inserted.user_id
JOIN feeds fd ON fd.id = inserted.feed_id;


-- name: GetFeedFollowsForUser :many
SELECT 
    ff.id,
    ff.user_id,
    ff.feed_id,
    ff.created_at,
    ff.updated_at,
    u.name AS user_name,
    fd.name AS feed_name
FROM feed_follows ff
JOIN users u ON u.id = ff.user_id
JOIN feeds fd ON fd.id = ff.feed_id
WHERE ff.user_id = $1;

