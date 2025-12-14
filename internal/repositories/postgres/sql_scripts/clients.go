package sql_scripts

const (
	InsertClient = `
		INSERT INTO clients
			(orbit_id, client_id, client_secret_hash, name, description,
			 redirect_uris, post_logout_redirect_uris, grant_types, response_types,
			 token_endpoint_auth_method, contacts, logo_uri, app_type,
			 is_public, is_active, allowed_cors_origins, allowed_scopes,
			 metadata, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,
			 $6,$7,$8,$9,
			 $10,$11,$12,$13,
			 $14,$15,$16,$17,
			 $18,$19,$20)
		RETURNING id, created_at, updated_at, deleted_at
	`

	SelectClientByID = `
		SELECT
			id, orbit_id, client_id, client_secret_hash, name, description,
			redirect_uris, post_logout_redirect_uris, grant_types, response_types,
			token_endpoint_auth_method, contacts, logo_uri, app_type,
			is_public, is_active, allowed_cors_origins, allowed_scopes,
			metadata, created_at, updated_at, deleted_at
		FROM clients
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	SelectClientByClientID = `
		SELECT
			id, orbit_id, client_id, client_secret_hash, name, description,
			redirect_uris, post_logout_redirect_uris, grant_types, response_types,
			token_endpoint_auth_method, contacts, logo_uri, app_type,
			is_public, is_active, allowed_cors_origins, allowed_scopes,
			metadata, created_at, updated_at, deleted_at
		FROM clients
		WHERE orbit_id = $1 AND client_id = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	UpdateClient = `
		UPDATE clients
		SET
			client_secret_hash = $3,
			name = $4,
			description = $5,
			redirect_uris = $6,
			post_logout_redirect_uris = $7,
			grant_types = $8,
			response_types = $9,
			token_endpoint_auth_method = $10,
			contacts = $11,
			logo_uri = $12,
			app_type = $13,
			is_public = $14,
			is_active = $15,
			allowed_cors_origins = $16,
			allowed_scopes = $17,
			metadata = $18,
			updated_at = $19
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, updated_at
	`

	SoftDeleteClient = `
		UPDATE clients
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`
)
