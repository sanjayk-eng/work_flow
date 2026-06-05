-- +goose Up
-- +goose StatementBegin

CREATE TYPE user_status AS ENUM (
    'pending',
    'active'
);

CREATE TABLE users (
    id UUID PRIMARY KEY,

    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,

    email_verified BOOLEAN NOT NULL DEFAULT FALSE,

    status user_status NOT NULL DEFAULT 'pending',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email
ON users(email);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_status;

-- +goose StatementEnd