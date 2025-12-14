package sql_scripts

const (
	InsertSession = `
		INSERT INTO sessions
			(orbit_id, user_id, client_id, started_at, last_active_at, expires_at, revoked, device_info, ip, metadata, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at, updated_at
	`

	SelectSessionByID = `
		SELECT id, orbit_id, user_id, client_id, started_at, last_active_at, expires_at, revoked, device_info, ip, metadata, created_at, updated_at
		FROM sessions
		WHERE id = $1
		LIMIT 1
	`

	UpdateSession = `
		UPDATE sessions
		SET last_active_at = $2, expires_at = $3, revoked = $4, metadata = $5, updated_at = $6
		WHERE id = $1
		RETURNING id, updated_at
	`

	RevokeSession = `
		UPDATE sessions
		SET revoked = TRUE, updated_at = $2
		WHERE id = $1
		RETURNING id
	`

	ListSessionsByUser = `
		SELECT id, orbit_id, user_id, client_id, started_at, last_active_at, expires_at, revoked, device_info, ip, metadata, created_at, updated_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`
)
