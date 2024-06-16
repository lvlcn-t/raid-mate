CREATE TABLE IF NOT EXISTS guilds (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    server_name TEXT NOT NULL,
    server_region TEXT NOT NULL,
    server_realm TEXT NOT NULL
);