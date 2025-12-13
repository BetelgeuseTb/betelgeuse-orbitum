CREATE TABLE orbitum.auth_codes
(
    id                    BIGSERIAL PRIMARY KEY,
    created_at            TIMESTAMPTZ  NOT NULL,
    expires_at            TIMESTAMPTZ  NOT NULL,
    code                  VARCHAR(512) NOT NULL UNIQUE,
    orbit_id              BIGINT       NOT NULL REFERENCES orbitum.orbits (id) ON DELETE CASCADE,
    client_id             BIGINT       NOT NULL REFERENCES orbitum.clients (id) ON DELETE CASCADE,
    user_id               BIGINT       REFERENCES orbitum.users (id) ON DELETE SET NULL,
    redirect_uri          TEXT         NOT NULL,
    scope                 JSONB,
    code_challenge        TEXT,
    code_challenge_method VARCHAR(10),
    used                  BOOLEAN      NOT NULL DEFAULT FALSE,
    metadata              JSONB
);

CREATE INDEX idx_auth_codes_expires_at
    ON orbitum.auth_codes (expires_at);
