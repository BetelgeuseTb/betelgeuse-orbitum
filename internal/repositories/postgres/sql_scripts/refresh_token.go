package sql_scripts

const (
	InsertRefreshToken = `
		INSERT INTO refresh_tokens
			(expires_at, token_string, jti, orbit_id, client_id, user_id, revoked, rotated_from_id, rotated_to_id, scopes, metadata, last_used_at, use_count, created_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, created_at, expires_at
	`

	SelectRefreshTokenByID = `
		SELECT id, expires_at, token_string, jti, orbit_id, client_id, user_id, revoked, rotated_from_id, rotated_to_id, scopes, metadata, last_used_at, use_count, created_at
		FROM refresh_tokens
		WHERE id = $1
		LIMIT 1
	`

	SelectRefreshTokenByJTI = `
		SELECT id, expires_at, token_string, jti, orbit_id, client_id, user_id, revoked, rotated_from_id, rotated_to_id, scopes, metadata, last_used_at, use_count, created_at
		FROM refresh_tokens
		WHERE jti = $1
		LIMIT 1
	`

	UpdateRefreshToken = `
		UPDATE refresh_tokens
		SET
			token_string = $2,
			revoked = $3,
			rotated_from_id = $4,
			rotated_to_id = $5,
			scopes = $6,
			metadata = $7,
			last_used_at = $8,
			use_count = $9,
			expires_at = $10
		WHERE id = $1
		RETURNING id, expires_at
	`

	RevokeRefreshTokenByJTI = `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE jti = $1 AND (revoked IS NULL OR revoked = FALSE)
		RETURNING id
	`

	RotateRefreshToken = `
		UPDATE refresh_tokens
		SET rotated_to_id = $2
		WHERE id = $1
		RETURNING id
	`
)
