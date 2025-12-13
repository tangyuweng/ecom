package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tangyuweng/ecom/internal/domain/repository"
)

type TokenRepository struct {
	client *redis.Client
}

func NewTokenRepository(client *redis.Client) repository.TokenRepository {
	return &TokenRepository{client: client}
}

// 將 token JTI 加入黑名單
func (r *TokenRepository) AddToBlacklist(ctx context.Context, jti string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:token:%s", jti)
	return r.client.Set(ctx, key, "1", ttl).Err()
}

// 檢查 token 是否在黑名單內
func (r *TokenRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("blacklist:token:%s", jti)
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// 取得用戶版本號
func (r *TokenRepository) GetUserVersion(ctx context.Context, userID string) (int, error) {
	key := fmt.Sprintf("user:version:%s", userID)
	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil // 如果不存在，返回版本 0
	}
	if err != nil {
		return 0, err
	}

	version, err := strconv.Atoi(result)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// 遞增用戶版本號，用於全局撤銷
func (r *TokenRepository) IncrementUserVersion(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:version:%s", userID)
	return r.client.Incr(ctx, key).Err()
}

// 設置用戶 token 版本號
func (r *TokenRepository) SetUserVersion(ctx context.Context, userID string, version int) error {
	key := fmt.Sprintf("user:version:%s", userID)
	return r.client.Set(ctx, key, version, 0).Err()
}
