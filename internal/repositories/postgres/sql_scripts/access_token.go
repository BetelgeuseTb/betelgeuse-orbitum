package sql_scripts

const (
	InsertAccessToken = `
		INSERT INTO access_tokens
			(jti, orbit_id, client_id, user_id, is_jwt, token_string, scope,
			 issued_at, token_type, revoked, metadata, refresh_token_id, created_at, expires_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, created_at, expires_at
	`

	SelectAccessTokenByID = `
		SELECT id, jti, orbit_id, client_id, user_id, is_jwt, token_string, scope,
		       issued_at, token_type, revoked, metadata, refresh_token_id, created_at, expires_at
		FROM access_tokens
		WHERE id = $1
		LIMIT 1
	`

	SelectAccessTokenByJTI = `
		SELECT id, jti, orbit_id, client_id, user_id, is_jwt, token_string, scope,
		       issued_at, token_type, revoked, metadata, refresh_token_id, created_at, expires_at
		FROM access_tokens
		WHERE jti = $1
		LIMIT 1
	`

	UpdateAccessToken = `
		UPDATE access_tokens
		SET
			token_string = $2,
			revoked = $3,
			metadata = $4,
			refresh_token_id = $5,
			issued_at = $6,
			expires_at = $7
		WHERE id = $1
		RETURNING id, jti, orbit_id, client_id, user_id, is_jwt, token_string, scope,
		       issued_at, token_type, revoked, metadata, refresh_token_id, created_at, expires_at
	`

	RevokeAccessTokenByJTI = `
		UPDATE access_tokens
		SET revoked = TRUE
		WHERE jti = $1 AND (revoked IS NULL OR revoked = FALSE)
		RETURNING id, revoked
	`
)
