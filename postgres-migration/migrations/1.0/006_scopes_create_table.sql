CREATE TABLE orbitum.scopes
(
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ  NOT NULL,
    updated_at  TIMESTAMPTZ  NOT NULL,
    orbit_id    BIGINT       NOT NULL REFERENCES orbitum.orbits (id) ON DELETE CASCADE,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    is_default  BOOLEAN      NOT NULL DEFAULT FALSE,
    is_required BOOLEAN      NOT NULL DEFAULT FALSE,
    UNIQUE (orbit_id, name)
);
