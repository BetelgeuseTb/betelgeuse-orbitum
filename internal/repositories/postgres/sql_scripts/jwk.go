package sql_scripts

const (
	InsertJWK = `
		INSERT INTO jwks
			(orbit_id, kid, "use", alg, kty, public_key_jwk, private_key_cipher, is_active, not_before, expires_at, metadata, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, created_at, updated_at
	`

	SelectJWKByID = `
		SELECT id, orbit_id, kid, "use", alg, kty, public_key_jwk, private_key_cipher, is_active, not_before, expires_at, metadata, created_at, updated_at
		FROM jwks
		WHERE id = $1
		LIMIT 1
	`

	SelectJWKByOrbitAndKid = `
		SELECT id, orbit_id, kid, "use", alg, kty, public_key_jwk, private_key_cipher, is_active, not_before, expires_at, metadata, created_at, updated_at
		FROM jwks
		WHERE orbit_id = $1 AND kid = $2
		LIMIT 1
	`

	UpdateJWK = `
		UPDATE jwks
		SET
			"public_key_jwk" = $3,
			"private_key_cipher" = $4,
			"is_active" = $5,
			"not_before" = $6,
			"expires_at" = $7,
			"metadata" = $8,
			"updated_at" = $9
		WHERE id = $1 AND orbit_id = $2
		RETURNING id, updated_at
	`

	ListJWKsByOrbit = `
		SELECT id, orbit_id, kid, "use", alg, kty, public_key_jwk, private_key_cipher, is_active, not_before, expires_at, metadata, created_at, updated_at
		FROM jwks
		WHERE orbit_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`
)
