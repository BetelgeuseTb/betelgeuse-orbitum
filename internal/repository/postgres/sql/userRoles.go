package sql

const (
	InsertUserRole = `
        INSERT INTO orbitum.user_roles (user_id, role_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING;
    `

	DeleteUserRole = `
        DELETE FROM orbitum.user_roles
        WHERE user_id = $1 AND role_id = $2;
    `

	DeleteAllUserRoles = `
        DELETE FROM orbitum.user_roles
        WHERE user_id = $1;
    `

	GetUserRoles = `
        SELECT role_id, assigned_at
        FROM orbitum.user_roles
        WHERE user_id = $1
        ORDER BY assigned_at ASC;
    `

	CheckUserHasRole = `
        SELECT 1
        FROM orbitum.user_roles
        WHERE user_id = $1 AND role_id = $2
        LIMIT 1;
    `
)
