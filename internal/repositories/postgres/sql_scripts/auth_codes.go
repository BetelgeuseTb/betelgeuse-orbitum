package sql_scripts

const (
	InsertAuthCode = `
		INSERT INTO auth_codes
			(code, orbit_id, client_id, user_id, redirect_uri, scope, code_challenge, code_challenge_method, used, metadata, created_at, expires_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, expires_at`

	SelectAuthCodeByID = `
		SELECT id, code, orbit_id, client_id, user_id, redirect_uri, scope,
		       code_challenge, code_challenge_method, used, metadata, created_at, expires_at
		FROM auth_codes
		WHERE id = $1
		LIMIT 1`

	SelectAuthCodeByCode = `
		SELECT id, code, orbit_id, client_id, user_id, redirect_uri, scope,
		       code_challenge, code_challenge_method, used, metadata, created_at, expires_at
		FROM auth_codes
		WHERE code = $1
		LIMIT 1`

	UpdateAuthCodeSetUsedByCode = `
		UPDATE auth_codes
		SET used = TRUE
		WHERE code = $1 AND used = FALSE
		RETURNING id`

	SoftDeleteAuthCode = `
        UPDATE auth_codes
        SET deleted_at = $2
        WHERE id = $1 AND deleted_at IS NULL
        RETURNING id`

	UpdateAuthCode = `
        UPDATE orbitum.auth_codes 
        SET redirect_uri = $2, scope = $3, metadata = $4, expires_at = $5
        WHERE id = $1
        RETURNING id`
)
