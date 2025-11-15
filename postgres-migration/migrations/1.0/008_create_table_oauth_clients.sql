CREATE TABLE orbitum.oauth_clients (
    client_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_secret TEXT,
    client_name TEXT NOT NULL,
    redirect_uris TEXT[] NOT NULL DEFAULT '{}',
    scopes TEXT[] NOT NULL DEFAULT '{}',
    grant_types TEXT[] NOT NULL,
    token_endpoint_auth_method TEXT NOT NULL DEFAULT 'client_secret_basic',
    public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_oauth_clients_client_id ON orbitum.oauth_clients(client_id);