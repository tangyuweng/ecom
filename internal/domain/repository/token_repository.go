package repository

import (
	"context"
	"time"
)

type TokenRepository interface {
	AddToBlacklist(ctx context.Context, jti string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)

	GetUserVersion(ctx context.Context, userID string) (int, error)
	IncrementUserVersion(ctx context.Context, userID string) error
	SetUserVersion(ctx context.Context, userID string, version int) error
}
