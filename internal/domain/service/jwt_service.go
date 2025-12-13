package service

import (
	"context"
	"time"
)

type TokenDetails struct {
	UserID    string
	JTI       string
	Version   int
	ExpiresAt time.Time
}

type JWTService interface {
	GenerateAccessToken(userID string, version int) (string, error)
	GenerateRefreshToken(userID string, version int) (string, error)
	ValidateToken(ctx context.Context, token string) (*TokenDetails, error)
	ValidateTokenBasic(token string) (*TokenDetails, error)
}
