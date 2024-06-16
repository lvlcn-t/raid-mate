-- name: NewGuild :exec
INSERT INTO guilds (
        id,
        name,
        server_name,
        server_region,
        server_realm
    )
VALUES ($1, $2, $3, $4, $5);

-- name: ListGuilds :many
SELECT id,
    name,
    server_name,
    server_region,
    server_realm
FROM guilds;

-- name: GetGuild :one
SELECT id,
    name,
    server_name,
    server_region,
    server_realm
FROM guilds
WHERE id = $1;

-- name: UpdateGuild :exec
UPDATE guilds
SET name = $1,
    server_name = $2,
    server_region = $3,
    server_realm = $4
WHERE id = $5
RETURNING *;

-- name: DeleteGuild :exec
DELETE FROM guilds
WHERE id = $1;

-- name: CountGuilds :one
SELECT COUNT(*)
FROM guilds;

-- name: FuzzyGuildSearch :many
SELECT *
FROM guilds
WHERE similarity(name, $1) > 0.15;