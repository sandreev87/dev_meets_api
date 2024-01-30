
CREATE TABLE IF NOT EXISTS users
(
    id        SERIAL PRIMARY KEY,
    email     TEXT NOT NULL UNIQUE,
    pass_hash TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);