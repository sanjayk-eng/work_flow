-- +goose Up
-- +goose StatementBegin

CREATE TABLE email_verifications (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    token TEXT NOT NULL UNIQUE,

    expires_at TIMESTAMPTZ NOT NULL,

    verified_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_email_verifications_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_email_verifications_user_id
ON email_verifications(user_id);

CREATE INDEX idx_email_verifications_token
ON email_verifications(token);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS email_verifications;

-- +goose StatementEnd