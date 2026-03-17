-- name: UpsertIntegration :one
INSERT INTO integrations (user_id, provider, status, access_token, refresh_token, token_expiry, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id, provider) DO UPDATE SET
    status = EXCLUDED.status,
    access_token = EXCLUDED.access_token,
    refresh_token = EXCLUDED.refresh_token,
    token_expiry = EXCLUDED.token_expiry,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
RETURNING *;

-- name: GetIntegrationsByUserID :many
SELECT * FROM integrations WHERE user_id = $1 ORDER BY provider;

-- name: GetIntegration :one
SELECT * FROM integrations WHERE user_id = $1 AND provider = $2;

-- name: DeleteIntegration :exec
DELETE FROM integrations WHERE user_id = $1 AND provider = $2;
