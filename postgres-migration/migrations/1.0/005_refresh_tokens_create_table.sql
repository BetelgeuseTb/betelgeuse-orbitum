CREATE TABLE orbitum.refresh_tokens
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ         NOT NULL,
    expires_at      TIMESTAMPTZ         NOT NULL,
    token_string    TEXT                NOT NULL,
    jti             VARCHAR(200) UNIQUE NOT NULL,
    orbit_id        BIGINT              NOT NULL REFERENCES orbitum.orbits (id),
    client_id       BIGINT              NOT NULL REFERENCES orbitum.clients (id),
    user_id         BIGINT,
    revoked         BOOLEAN,
    rotated_from_id BIGINT,
    rotated_to_id   BIGINT,
    scopes          JSONB,
    metadata        JSONB,
    last_used_at    TIMESTAMPTZ,
    use_count       INT
);
