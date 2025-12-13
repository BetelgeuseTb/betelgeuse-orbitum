CREATE TABLE orbitum.jwks
(
    id                 BIGSERIAL PRIMARY KEY,
    created_at         TIMESTAMPTZ  NOT NULL,
    updated_at         TIMESTAMPTZ  NOT NULL,
    orbit_id           BIGINT       NOT NULL REFERENCES orbitum.orbits (id) ON DELETE CASCADE,
    kid                VARCHAR(200) NOT NULL,
    use                VARCHAR(50),
    alg                VARCHAR(50),
    kty                VARCHAR(50),
    public_key_jwk     JSONB        NOT NULL,
    private_key_cipher TEXT,
    is_active          BOOLEAN      NOT NULL DEFAULT FALSE,
    not_before         TIMESTAMPTZ,
    expires_at         TIMESTAMPTZ,
    metadata           JSONB,
    UNIQUE (orbit_id, kid)
);

CREATE INDEX idx_jwks_active
    ON orbitum.jwks (orbit_id, is_active);
