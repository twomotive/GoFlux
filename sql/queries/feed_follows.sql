-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT 
    ff.id,
    ff.created_at,
    ff.updated_at,
    ff.user_id,
    ff.feed_id,
    u.name AS user_name,
    f.name AS feed_name
FROM inserted_feed_follow ff
JOIN users u ON ff.user_id = u.id
JOIN feeds f ON ff.feed_id = f.id;


-- name: GetFeedFollowsByUser :many
SELECT 
    ff.id,
    ff.created_at,
    ff.updated_at,
    ff.user_id,
    ff.feed_id,
    u.name AS user_name,
    f.name AS feed_name,
    f.url AS feed_url
FROM feed_follows ff
JOIN users u ON ff.user_id = u.id
JOIN feeds f ON ff.feed_id = f.id
WHERE ff.user_id = $1;


-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE
    feed_follows.user_id = $1
    AND feed_id IN (
        SELECT id FROM feeds WHERE url = $2
    );