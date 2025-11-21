package sql

const(

    RevokedTokenInsert = `
        INSERT INTO orbitum.revoked_tokens (jti, token_type, revoked_at, reason)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (jti) DO NOTHING;
    `

    RevokedTokenCheck = `
        SELECT 1
        FROM orbitum.revoked_tokens
        WHERE jti = $1
        LIMIT 1;
    `
) 