CREATE TABLE orbitum.totps
(
    id             BIGSERIAL PRIMARY KEY,
    created_at     TIMESTAMPTZ NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL,
    user_id        BIGINT      NOT NULL REFERENCES orbitum.users (id) ON DELETE CASCADE,
    orbit_id       BIGINT      NOT NULL REFERENCES orbitum.orbits (id) ON DELETE CASCADE,
    secret_cipher  TEXT        NOT NULL,
    algorithm      VARCHAR(20) NOT NULL,
    digits         INT         NOT NULL,
    period INT NOT NULL,
    issuer         VARCHAR(200),
    label          VARCHAR(200),
    last_used_step BIGINT      NOT NULL DEFAULT 0,
    is_confirmed   BOOLEAN     NOT NULL DEFAULT FALSE,
    name           VARCHAR(100)
);
