-- name: GetCredentials :one
SELECT id,
    guild_id,
    name,
    url,
    username,
    password
FROM credentials
WHERE guild_id = $1
    AND name = $2;

-- name: SetCredentials :exec
INSERT INTO credentials (guild_id, name, url, username, password)
VALUES ($1, $2, $3, $4, $5) ON CONFLICT (guild_id, name) DO
UPDATE
SET url = EXCLUDED.url,
    username = EXCLUDED.username,
    password = EXCLUDED.password
RETURNING *;