package sql_scripts

const (
	InsertScope = `
		INSERT INTO scopes
			(orbit_id, name, description, is_default, is_active, metadata, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at, deleted_at
	`

	SelectScopeByID = `
		SELECT
			id, orbit_id, name, description, is_default, is_active,
			metadata, created_at, updated_at, deleted_at
		FROM scopes
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	SelectScopeByName = `
		SELECT
			id, orbit_id, name, description, is_default, is_active,
			metadata, created_at, updated_at, deleted_at
		FROM scopes
		WHERE orbit_id = $1 AND name = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	ListScopesByOrbit = `
		SELECT
			id, orbit_id, name, description, is_default, is_active,
			metadata, created_at, updated_at, deleted_at
		FROM scopes
		WHERE orbit_id = $1 AND deleted_at IS NULL
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`

	UpdateScope = `
		UPDATE scopes
		SET
			description = $3,
			is_default = $4,
			is_active = $5,
			metadata = $6,
			updated_at = $7
		WHERE id = $1 AND orbit_id = $2 AND deleted_at IS NULL
		RETURNING id, updated_at
	`

	SoftDeleteScope = `
		UPDATE scopes
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`
)
