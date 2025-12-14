package sql_scripts

const (
	InsertAuditLog = `
		INSERT INTO audit_logs
			(created_at, actor_user_id, actor_client_id, action, result, ip, orbit_id, details)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, created_at
	`

	ListAuditLogsByOrbit = `
		SELECT id, created_at, actor_user_id, actor_client_id, action, result, ip, orbit_id, details
		FROM audit_logs
		WHERE orbit_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`
)
