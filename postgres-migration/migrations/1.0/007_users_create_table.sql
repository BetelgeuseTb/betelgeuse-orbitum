CREATE TABLE orbitum.users
(
    id                   BIGSERIAL PRIMARY KEY,
    created_at           TIMESTAMPTZ  NOT NULL,
    updated_at           TIMESTAMPTZ  NOT NULL,
    deleted_at           TIMESTAMPTZ,
    orbit_id             BIGINT       NOT NULL REFERENCES orbitum.orbits (id) ON DELETE CASCADE,
    username             VARCHAR(200) NOT NULL,
    email                VARCHAR(255),
    email_verified       BOOLEAN      NOT NULL DEFAULT FALSE,
    password_hash        VARCHAR(512) NOT NULL,
    password_algo        VARCHAR(50)  NOT NULL,
    last_password_change TIMESTAMPTZ,
    display_name         VARCHAR(255),
    profile              JSONB,
    is_active            BOOLEAN      NOT NULL DEFAULT TRUE,
    is_locked            BOOLEAN      NOT NULL DEFAULT FALSE,
    mfa_enabled          BOOLEAN      NOT NULL DEFAULT FALSE,
    metadata             JSONB,
    UNIQUE (orbit_id, username)
);
