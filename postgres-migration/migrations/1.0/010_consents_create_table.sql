CREATE TABLE orbitum.consents
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    orbit_id   BIGINT      NOT NULL REFERENCES orbitum.orbits (id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES orbitum.users (id) ON DELETE CASCADE,
    client_id  BIGINT      NOT NULL REFERENCES orbitum.clients (id) ON DELETE CASCADE,
    scopes     JSONB       NOT NULL,
    granted_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    revoked    BOOLEAN     NOT NULL DEFAULT FALSE,
    UNIQUE (user_id, client_id)
);
