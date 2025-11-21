package sql

const (
	AccessTokenInsert = `
		INSERT INTO orbitum.access_tokens
			(token_id, client_id, user_id, scopes, issued_at, expires_at, jti)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	AccessTokenGetByJTI = `
		SELECT token_id, client_id, user_id, scopes, issued_at, expires_at, jti
		FROM orbitum.access_tokens
		WHERE jti = $1
		LIMIT 1
	`

	AccessTokenDeleteExpired = `
		DELETE FROM orbitum.access_tokens
		WHERE expires_at < NOW()
	`
)
