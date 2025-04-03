-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;


-- name: GetFeedsWithUserNames :many
SELECT 
    f.name AS feed_name,
    f.url AS feed_url,
    u.name AS user_name
FROM 
    feeds f
JOIN 
    users u ON f.user_id = u.id;



-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;