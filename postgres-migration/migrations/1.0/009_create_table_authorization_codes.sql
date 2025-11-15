CREATE TABLE orbitum.authorization_codes (
    code TEXT PRIMARY KEY,
    client_id UUID NOT NULL REFERENCES orbitum.oauth_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES orbitum.users(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL,
    redirect_uri TEXT NOT NULL,
    code_challenge TEXT,
    code_challenge_method TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_authorization_codes_client_id ON orbitum.authorization_codes(client_id);
CREATE INDEX idx_authorization_codes_user_id ON orbitum.authorization_codes(user_id);
CREATE INDEX idx_authorization_codes_expires ON orbitum.authorization_codes(expires_at);