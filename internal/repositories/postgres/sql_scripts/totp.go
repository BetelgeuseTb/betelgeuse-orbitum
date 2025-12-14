package sql_scripts

const (
	InsertTOTP = `
		INSERT INTO totps
			(created_at, updated_at, user_id, orbit_id, secret_cipher, algorithm, digits, period, issuer, label, last_used_step, is_confirmed, name)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, created_at, updated_at
	`

	SelectTOTPByID = `
		SELECT id, user_id, orbit_id, secret_cipher, algorithm, digits, period, issuer, label, last_used_step, is_confirmed, name, created_at, updated_at
		FROM totps
		WHERE id = $1
		LIMIT 1
	`

	UpdateTOTP = `
		UPDATE totps
		SET secret_cipher = $2, algorithm = $3, digits = $4, period = $5, issuer = $6, label = $7, last_used_step = $8, is_confirmed = $9, name = $10, updated_at = $11
		WHERE id = $1
		RETURNING id, updated_at
	`

	DeleteTOTP = `
		DELETE FROM totps
		WHERE id = $1
		RETURNING id
	`

	ListTOTPsByUser = `
		SELECT id, user_id, orbit_id, secret_cipher, algorithm, digits, period, issuer, label, last_used_step, is_confirmed, name, created_at, updated_at
		FROM totps
		WHERE user_id = $1
		ORDER BY id ASC
	`
)
