package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestTokenRepository_AddToBlacklist(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	repo := NewTokenRepository(client)
	ctx := context.Background()

	t.Run("成功將 token 加入黑名單", func(t *testing.T) {
		jti := "test-jti-123"
		ttl := 1 * time.Hour

		err := repo.AddToBlacklist(ctx, jti, ttl)
		assert.NoError(t, err)

		// 驗證 key 存在
		exists, err := client.Exists(ctx, "blacklist:token:"+jti).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	})

	t.Run("設置正確的 TTL", func(t *testing.T) {
		jti := "test-jti-456"
		ttl := 2 * time.Hour

		err := repo.AddToBlacklist(ctx, jti, ttl)
		assert.NoError(t, err)

		// 驗證 TTL
		result, err := client.TTL(ctx, "blacklist:token:"+jti).Result()
		assert.NoError(t, err)
		assert.True(t, result > 0)
		assert.True(t, result <= ttl)
	})
}

func TestTokenRepository_IsBlacklisted(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	repo := NewTokenRepository(client)
	ctx := context.Background()

	t.Run("檢查存在於黑名單的 token", func(t *testing.T) {
		jti := "blacklisted-token"
		err := client.Set(ctx, "blacklist:token:"+jti, "1", 1*time.Hour).Err()
		require.NoError(t, err)

		isBlacklisted, err := repo.IsBlacklisted(ctx, jti)
		assert.NoError(t, err)
		assert.True(t, isBlacklisted)
	})

	t.Run("檢查不存在於黑名單的 token", func(t *testing.T) {
		jti := "valid-token"

		isBlacklisted, err := repo.IsBlacklisted(ctx, jti)
		assert.NoError(t, err)
		assert.False(t, isBlacklisted)
	})
}

func TestTokenRepository_GetUserVersion(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	repo := NewTokenRepository(client)
	ctx := context.Background()

	t.Run("獲取已存在的用戶版本", func(t *testing.T) {
		userID := "user-123"
		expectedVersion := 5

		err := client.Set(ctx, "user:version:"+userID, expectedVersion, 0).Err()
		require.NoError(t, err)

		version, err := repo.GetUserVersion(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedVersion, version)
	})

	t.Run("獲取不存在的用戶版本應返回 0", func(t *testing.T) {
		userID := "new-user"

		version, err := repo.GetUserVersion(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, 0, version)
	})
}

func TestTokenRepository_SetUserVersion(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	repo := NewTokenRepository(client)
	ctx := context.Background()

	t.Run("成功設置用戶版本", func(t *testing.T) {
		userID := "user-456"
		version := 3

		err := repo.SetUserVersion(ctx, userID, version)
		assert.NoError(t, err)

		// 驗證值已設置
		result, err := client.Get(ctx, "user:version:"+userID).Result()
		assert.NoError(t, err)
		assert.Equal(t, "3", result)
	})

	t.Run("覆蓋已存在的版本", func(t *testing.T) {
		userID := "user-789"

		err := repo.SetUserVersion(ctx, userID, 1)
		require.NoError(t, err)

		err = repo.SetUserVersion(ctx, userID, 10)
		assert.NoError(t, err)

		version, err := repo.GetUserVersion(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, 10, version)
	})
}

func TestTokenRepository_IncrementUserVersion(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	repo := NewTokenRepository(client)
	ctx := context.Background()

	t.Run("遞增已存在的版本號", func(t *testing.T) {
		userID := "user-increment-1"

		err := repo.SetUserVersion(ctx, userID, 5)
		require.NoError(t, err)

		err = repo.IncrementUserVersion(ctx, userID)
		assert.NoError(t, err)

		version, err := repo.GetUserVersion(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, 6, version)
	})

	t.Run("遞增不存在的版本號（從 0 開始）", func(t *testing.T) {
		userID := "user-increment-2"

		err := repo.IncrementUserVersion(ctx, userID)
		assert.NoError(t, err)

		version, err := repo.GetUserVersion(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, 1, version)
	})

	t.Run("多次遞增", func(t *testing.T) {
		userID := "user-increment-3"

		for i := 0; i < 5; i++ {
			err := repo.IncrementUserVersion(ctx, userID)
			assert.NoError(t, err)
		}

		version, err := repo.GetUserVersion(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, 5, version)
	})
}
