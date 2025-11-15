CREATE TABLE orbitum.access_tokens (
    token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES orbitum.oauth_clients(client_id) ON DELETE CASCADE,
    user_id UUID REFERENCES orbitum.users(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    jti UUID NOT NULL DEFAULT gen_random_uuid()
);

CREATE INDEX idx_access_tokens_client_id ON orbitum.access_tokens(client_id);
CREATE INDEX idx_access_tokens_user_id ON orbitum.access_tokens(user_id);
CREATE INDEX idx_access_tokens_expires ON orbitum.access_tokens(expires_at);
CREATE INDEX idx_access_tokens_jti ON orbitum.access_tokens(jti);