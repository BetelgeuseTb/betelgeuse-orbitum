CREATE INDEX idx_refresh_tokens_active
    ON orbitum.refresh_tokens (jti)
    WHERE revoked = FALSE;

CREATE INDEX idx_refresh_tokens_expires
    ON orbitum.refresh_tokens (expires_at);
