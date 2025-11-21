package sql

const (
	OAuthClientGetByID = `
		SELECT client_id,
			client_secret_hash,
			client_name,
			redirect_uris,
			scopes,
			grant_types,
			token_endpoint_auth_method,
			public,
			created_at,
			updated_at
		FROM oauth_clients
		WHERE client_id = $1;
	`

	OAuthClientCreate = `
		INSERT INTO oauth_clients (
			client_id,
			client_secret_hash,
			client_name,
			redirect_uris,
			scopes,
			grant_types,
			token_endpoint_auth_method,
			public
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`

	OAuthClientUpdate = `
		UPDATE oauth_clients
		SET client_secret_hash = $1,
			client_name = $2,
			redirect_uris = $3,
			scopes = $4,
			grant_types = $5,
			token_endpoint_auth_method = $6,
			public = $7,
			updated_at = NOW()
		WHERE client_id = $8;
	`

	OAuthClientDelete = `
		DELETE FROM oauth_clients
		WHERE client_id = $1;
	`
)
