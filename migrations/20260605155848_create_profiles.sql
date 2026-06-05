-- +goose Up
-- +goose StatementBegin

CREATE TABLE profiles (
    user_id UUID PRIMARY KEY,

    first_name VARCHAR(100),
    last_name VARCHAR(100),

    avatar_url TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_profiles_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS profiles;

-- +goose StatementEnd