CREATE TABLE orbitum.sessions
(
    id             BIGSERIAL PRIMARY KEY,
    created_at     TIMESTAMPTZ NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL,
    orbit_id       BIGINT      NOT NULL REFERENCES orbitum.orbits (id) ON DELETE CASCADE,
    user_id        BIGINT      NOT NULL REFERENCES orbitum.users (id) ON DELETE CASCADE,
    client_id      BIGINT      REFERENCES orbitum.clients (id) ON DELETE SET NULL,
    started_at     TIMESTAMPTZ NOT NULL,
    last_active_at TIMESTAMPTZ NOT NULL,
    expires_at     TIMESTAMPTZ,
    revoked        BOOLEAN     NOT NULL DEFAULT FALSE,
    device_info    TEXT,
    ip             VARCHAR(100),
    metadata       JSONB
);

CREATE INDEX idx_sessions_active
    ON orbitum.sessions (user_id, revoked);
