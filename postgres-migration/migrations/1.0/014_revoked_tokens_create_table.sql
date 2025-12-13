CREATE TABLE orbitum.revoked_tokens
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ  NOT NULL,
    jti        VARCHAR(200) NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    orbit_id   BIGINT       NOT NULL REFERENCES orbitum.orbits (id) ON DELETE CASCADE,
    reason     VARCHAR(255)
);

CREATE INDEX idx_revoked_tokens_jti
    ON orbitum.revoked_tokens (jti);

CREATE INDEX idx_revoked_tokens_expires
    ON orbitum.revoked_tokens (expires_at);
