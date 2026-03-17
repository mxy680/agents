package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID         pgtype.UUID        `json:"id"`
	GoogleID   string             `json:"google_id"`
	Email      string             `json:"email"`
	Name       string             `json:"name"`
	PictureURL string             `json:"picture_url"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `json:"updated_at"`
}

type Integration struct {
	ID           pgtype.UUID        `json:"id"`
	UserID       pgtype.UUID        `json:"user_id"`
	Provider     string             `json:"provider"`
	Status       string             `json:"status"`
	AccessToken  pgtype.Text        `json:"access_token"`
	RefreshToken pgtype.Text        `json:"refresh_token"`
	TokenExpiry  pgtype.Timestamptz `json:"token_expiry"`
	Metadata     []byte             `json:"metadata"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
}
