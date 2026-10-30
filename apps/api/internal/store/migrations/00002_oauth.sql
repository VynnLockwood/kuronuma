-- +goose Up
-- OAuth authorization server for MCP clients. Codes and tokens are stored only as SHA-256 hashes.
CREATE TABLE oauth_clients (
    client_id     text PRIMARY KEY,
    client_name   text NOT NULL,
    redirect_uris text[] NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE oauth_codes (
    code_hash      bytea PRIMARY KEY,
    client_id      text NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redirect_uri   text NOT NULL,
    code_challenge text NOT NULL,
    scopes         text[] NOT NULL,
    resource       text NOT NULL,
    expires_at     timestamptz NOT NULL
);

CREATE INDEX oauth_codes_expires_at_idx ON oauth_codes (expires_at);

CREATE TABLE oauth_tokens (
    token_hash bytea PRIMARY KEY,
    kind       text NOT NULL CHECK (kind IN ('access', 'refresh')),
    client_id  text NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scopes     text[] NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX oauth_tokens_user_id_idx ON oauth_tokens (user_id);
CREATE INDEX oauth_tokens_expires_at_idx ON oauth_tokens (expires_at);

-- +goose Down
DROP TABLE oauth_tokens;
DROP TABLE oauth_codes;
DROP TABLE oauth_clients;
