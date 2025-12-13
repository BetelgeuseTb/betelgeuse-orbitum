CREATE TABLE orbitum.access_tokens
(
    id               BIGSERIAL PRIMARY KEY,
    created_at       TIMESTAMPTZ         NOT NULL,
    expires_at       TIMESTAMPTZ         NOT NULL,
    jti              VARCHAR(200) UNIQUE NOT NULL,
    orbit_id         BIGINT              NOT NULL REFERENCES orbitum.orbits (id),
    client_id        BIGINT              NOT NULL REFERENCES orbitum.clients (id),
    user_id          BIGINT,
    is_jwt           BOOLEAN,
    token_string     TEXT,
    scope            JSONB,
    issued_at        TIMESTAMPTZ,
    token_type       VARCHAR(50),
    revoked          BOOLEAN,
    metadata         JSONB,
    refresh_token_id BIGINT
);
