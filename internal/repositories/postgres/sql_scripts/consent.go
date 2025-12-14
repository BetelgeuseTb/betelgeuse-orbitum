package sql_scripts

const (
	InsertConsent = `
		INSERT INTO consents
			(created_at, updated_at, orbit_id, user_id, client_id, scopes, granted_at, expires_at, revoked)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, updated_at
	`

	SelectConsentByID = `
		SELECT
			id, orbit_id, user_id, client_id, scopes, granted_at, expires_at, revoked, created_at, updated_at
		FROM consents
		WHERE id = $1
		LIMIT 1
	`

	SelectConsentByUserAndClient = `
		SELECT
			id, orbit_id, user_id, client_id, scopes, granted_at, expires_at, revoked, created_at, updated_at
		FROM consents
		WHERE user_id = $1 AND client_id = $2
		LIMIT 1
	`

	UpdateConsent = `
		UPDATE consents
		SET scopes = $2, granted_at = $3, expires_at = $4, revoked = $5, updated_at = $6
		WHERE id = $1
		RETURNING id, updated_at
	`

	RevokeConsent = `
		UPDATE consents
		SET revoked = TRUE, updated_at = $2
		WHERE id = $1
		RETURNING id
	`

	ListConsentsByOrbit = `
		SELECT
			id, orbit_id, user_id, client_id, scopes, granted_at, expires_at, revoked, created_at, updated_at
		FROM consents
		WHERE orbit_id = $1
		ORDER BY id DESC
	`
)
