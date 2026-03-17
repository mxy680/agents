package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const upsertIntegration = `
INSERT INTO integrations (user_id, provider, status, access_token, refresh_token, token_expiry, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id, provider) DO UPDATE SET
    status = EXCLUDED.status,
    access_token = EXCLUDED.access_token,
    refresh_token = EXCLUDED.refresh_token,
    token_expiry = EXCLUDED.token_expiry,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
RETURNING id, user_id, provider, status, access_token, refresh_token, token_expiry, metadata, created_at, updated_at
`

type UpsertIntegrationParams struct {
	UserID       pgtype.UUID        `json:"user_id"`
	Provider     string             `json:"provider"`
	Status       string             `json:"status"`
	AccessToken  pgtype.Text        `json:"access_token"`
	RefreshToken pgtype.Text        `json:"refresh_token"`
	TokenExpiry  pgtype.Timestamptz `json:"token_expiry"`
	Metadata     []byte             `json:"metadata"`
}

func (q *Queries) UpsertIntegration(ctx context.Context, arg UpsertIntegrationParams) (Integration, error) {
	row := q.db.QueryRow(ctx, upsertIntegration,
		arg.UserID,
		arg.Provider,
		arg.Status,
		arg.AccessToken,
		arg.RefreshToken,
		arg.TokenExpiry,
		arg.Metadata,
	)
	var i Integration
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Provider,
		&i.Status,
		&i.AccessToken,
		&i.RefreshToken,
		&i.TokenExpiry,
		&i.Metadata,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getIntegrationsByUserID = `
SELECT id, user_id, provider, status, access_token, refresh_token, token_expiry, metadata, created_at, updated_at
FROM integrations WHERE user_id = $1 ORDER BY provider
`

func (q *Queries) GetIntegrationsByUserID(ctx context.Context, userID pgtype.UUID) ([]Integration, error) {
	rows, err := q.db.Query(ctx, getIntegrationsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Integration{}
	for rows.Next() {
		var i Integration
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Provider,
			&i.Status,
			&i.AccessToken,
			&i.RefreshToken,
			&i.TokenExpiry,
			&i.Metadata,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const getIntegration = `
SELECT id, user_id, provider, status, access_token, refresh_token, token_expiry, metadata, created_at, updated_at
FROM integrations WHERE user_id = $1 AND provider = $2
`

func (q *Queries) GetIntegration(ctx context.Context, userID pgtype.UUID, provider string) (Integration, error) {
	row := q.db.QueryRow(ctx, getIntegration, userID, provider)
	var i Integration
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Provider,
		&i.Status,
		&i.AccessToken,
		&i.RefreshToken,
		&i.TokenExpiry,
		&i.Metadata,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteIntegration = `
DELETE FROM integrations WHERE user_id = $1 AND provider = $2
`

func (q *Queries) DeleteIntegration(ctx context.Context, userID pgtype.UUID, provider string) error {
	_, err := q.db.Exec(ctx, deleteIntegration, userID, provider)
	return err
}
