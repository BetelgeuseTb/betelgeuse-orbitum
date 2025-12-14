package sql_scripts

const (
	InsertRevokedToken = `
		INSERT INTO revoked_tokens
			(created_at, jti, expires_at, orbit_id, reason)
		VALUES
			($1,$2,$3,$4,$5)
		RETURNING id, created_at
	`

	SelectRevokedByJTI = `
		SELECT id, created_at, jti, expires_at, orbit_id, reason
		FROM revoked_tokens
		WHERE jti = $1
		LIMIT 1
	`

	CountRevokedByJTI = `
		SELECT COUNT(1) FROM revoked_tokens WHERE jti = $1
	`
)
