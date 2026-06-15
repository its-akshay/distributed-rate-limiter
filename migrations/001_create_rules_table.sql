CREATE TABLE IF NOT EXISTS rules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    limit_count INT NOT NULL CHECK (limit_count > 0),
    window_seconds INT NOT NULL CHECK (window_seconds > 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);