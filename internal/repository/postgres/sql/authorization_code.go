package sql

const (

	AuthorizationCodeCreate = `
		INSERT INTO authorization_codes (code, client_id, redirect_uri, expires_at, user_id)
		VALUES ($1, $2, $3, $4, $5);
	`

	AuthorizationCodeGet = `
		SELECT code, client_id, redirect_uri, expires_at, user_id, used
		FROM authorization_codes
		WHERE code = $1;
	`

	AuthorizationCodeMarkUsed = `
		UPDATE authorization_codes
		SET used = TRUE
		WHERE code = $1 AND used = FALSE;
	`

	AuthorizationCodeDeleteExpired = `
		DELETE FROM authorization_codes
		WHERE expires_at < $1;
	`
)
