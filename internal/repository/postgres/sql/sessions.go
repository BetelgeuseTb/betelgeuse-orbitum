package sql

const (
	SessionInsert = `
        INSERT INTO orbitum.sessions 
            (user_id, refresh_token_hash, user_agent, ip_address, expires_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, user_id, refresh_token_hash, user_agent, ip_address,
                expires_at, created_at, updated_at, revoked;
    `

	SessionGetByID = `
        SELECT id, user_id, refresh_token_hash, user_agent, ip_address,
            expires_at, created_at, updated_at, revoked
        FROM orbitum.sessions
        WHERE id = $1;
    `

	SessionGetByUser = `
        SELECT id, user_id, refresh_token_hash, user_agent, ip_address,
            expires_at, created_at, updated_at, revoked
        FROM orbitum.sessions
        WHERE user_id = $1
        ORDER BY created_at DESC;
    `

	SessionRevoke = `
        UPDATE orbitum.sessions
        SET revoked = TRUE, updated_at = NOW()
        WHERE id = $1;
    `

	SessionDeleteExpired = `
        DELETE FROM orbitum.sessions
        WHERE expires_at < NOW() OR revoked = TRUE;
    `
)
