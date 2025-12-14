package sql_scripts

const (
	InsertOrbit = `
		INSERT INTO orbits
			(name, display_name, description, issuer, domain, config, default_scopes, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING 
			id, created_at, updated_at, deleted_at
		`

	SelectOrbitByID = `
		SELECT 
			id, name, display_name, description, issuer, domain, config, default_scopes, 
			created_at, updated_at, deleted_at
		FROM orbits
		WHERE 
			id = $1 
			AND deleted_at IS NULL
		LIMIT 1
		`

	SelectOrbitByName = `
		SELECT 
			id, name, display_name, description, issuer, domain, config, default_scopes, 
			created_at, updated_at, deleted_at
		FROM orbits
		WHERE 
			name = $1 
			AND deleted_at IS NULL
		LIMIT 1
		`

	UpdateOrbit = `
		UPDATE orbits
		SET
			name         = $2,
			display_name = $3,
			description  = $4,
			issuer       = $5,
			domain       = $6,
			config       = $7,
			default_scopes = $8,
			updated_at   = $9
		WHERE 
			id = $1 
			AND deleted_at IS NULL
		RETURNING 
			id, updated_at
		`

	SoftDeleteOrbit = `
		UPDATE orbits
		SET 
			deleted_at = $2, 
			updated_at = $2
		WHERE 
			id = $1 
			AND deleted_at IS NULL
		RETURNING id
		`

	ListOrbits = `
		SELECT 
			id, name, display_name, description, issuer, domain, config, default_scopes, 
			created_at, updated_at, deleted_at
		FROM orbits
		WHERE deleted_at IS NULL
		ORDER BY id ASC
		LIMIT $1 OFFSET $2
		`
)
