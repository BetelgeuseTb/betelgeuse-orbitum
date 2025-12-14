package sql_scripts

const (
	InsertUser = `
		INSERT INTO users
			(orbit_id, username, email, email_verified, password_hash, password_algo, last_password_change,
			 display_name, profile, is_active, is_locked, mfa_enabled, metadata, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7,
			 $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at, deleted_at
		`

	SelectUserByID = `
		SELECT id, orbit_id, username, email, email_verified, password_hash, password_algo, last_password_change,
			   display_name, profile, is_active, is_locked, mfa_enabled, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
		`

	SelectUserByEmail = `
		SELECT id, orbit_id, username, email, email_verified, password_hash, password_algo, last_password_change,
			   display_name, profile, is_active, is_locked, mfa_enabled, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE orbit_id = $1 AND lower(email) = lower($2) AND deleted_at IS NULL
		LIMIT 1
		`

	SelectUserByUsername = `
		SELECT id, orbit_id, username, email, email_verified, password_hash, password_algo, last_password_change,
			   display_name, profile, is_active, is_locked, mfa_enabled, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE orbit_id = $1 AND username = $2 AND deleted_at IS NULL
		LIMIT 1
		`

	UpdateUser = `
		UPDATE users
		SET
			username = $2,
			email = $3,
			email_verified = $4,
			password_hash = $5,
			password_algo = $6,
			last_password_change = $7,
			display_name = $8,
			profile = $9,
			is_active = $10,
			is_locked = $11,
			mfa_enabled = $12,
			metadata = $13,
			updated_at = $14
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, updated_at
		`

	SoftDeleteUser = `
		UPDATE users
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
		`

	ListUsersByOrbit = `
		SELECT id, orbit_id, username, email, email_verified, password_hash, password_algo, last_password_change,
			   display_name, profile, is_active, is_locked, mfa_enabled, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE orbit_id = $1 AND deleted_at IS NULL
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
		`
)
