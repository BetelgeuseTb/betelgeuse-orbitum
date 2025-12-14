CREATE TABLE orbitum.orbits
(
    id             BIGSERIAL PRIMARY KEY,
    created_at     TIMESTAMPTZ  NOT NULL,
    updated_at     TIMESTAMPTZ  NOT NULL,
    deleted_at     TIMESTAMPTZ,
    name           VARCHAR(200) NOT NULL UNIQUE,
    display_name   VARCHAR(255),
    description    TEXT,
    issuer         VARCHAR(512) NOT NULL,
    models         VARCHAR(255),
    config         JSONB,
    default_scopes JSONB
);
