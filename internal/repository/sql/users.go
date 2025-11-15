package sql

const (
	UserInsert = `
		INSERT INTO orbitum.users (email, password_hash, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, is_active, created_at, updated_at;
	`

	UserGetByID = `
		SELECT id, email, password_hash, is_active, created_at, updated_at
		FROM orbitum.users
		WHERE id = $1;
	`

	UserGetByEmail = `
		SELECT id, email, password_hash, is_active, created_at, updated_at
		FROM orbitum.users
		WHERE email = $1;
	`

	UserUpdate = `
		UPDATE orbitum.users
		SET email = $2, password_hash = $3, is_active = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, password_hash, is_active, created_at, updated_at;
	`

	UserDelete = `DELETE FROM orbitum.users WHERE id = $1;`
)
