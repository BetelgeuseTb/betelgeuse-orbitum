CREATE INDEX idx_access_tokens_active
    ON orbitum.access_tokens (jti)
    WHERE revoked = FALSE;

CREATE INDEX idx_access_tokens_expires
    ON orbitum.access_tokens (expires_at);
