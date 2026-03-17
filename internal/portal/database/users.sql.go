package database

import (
	"context"
)

const createUser = `
INSERT INTO users (google_id, email, name, picture_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (google_id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    picture_url = EXCLUDED.picture_url,
    updated_at = NOW()
RETURNING id, google_id, email, name, picture_url, created_at, updated_at
`

type CreateUserParams struct {
	GoogleID   string `json:"google_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	PictureURL string `json:"picture_url"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.GoogleID, arg.Email, arg.Name, arg.PictureURL)
	var i User
	err := row.Scan(
		&i.ID,
		&i.GoogleID,
		&i.Email,
		&i.Name,
		&i.PictureURL,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByGoogleID = `
SELECT id, google_id, email, name, picture_url, created_at, updated_at FROM users WHERE google_id = $1
`

func (q *Queries) GetUserByGoogleID(ctx context.Context, googleID string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByGoogleID, googleID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.GoogleID,
		&i.Email,
		&i.Name,
		&i.PictureURL,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByID = `
SELECT id, google_id, email, name, picture_url, created_at, updated_at FROM users WHERE id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id interface{}) (User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.GoogleID,
		&i.Email,
		&i.Name,
		&i.PictureURL,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
