package jwt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	repo "github.com/tangyuweng/ecom/internal/domain/repository"
)

func setupJWTService(tokenRepo *repo.MockTokenRepository) *JWTSvc {
	return &JWTSvc{
		secretKey:          "test-secret-key",
		accessTokenExpiry:  15 * time.Minute,
		refreshTokenExpiry: 24 * time.Hour,
		tokenRepo:          tokenRepo,
	}
}

func TestJWTSvc_GenerateAccessToken(t *testing.T) {
	mockTokenRepo := new(repo.MockTokenRepository)
	jwtSvc := setupJWTService(mockTokenRepo)

	t.Run("成功生成 access token", func(t *testing.T) {
		userID := "user-123"
		version := 1

		token, err := jwtSvc.GenerateAccessToken(userID, version)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// 驗證 token 可以被解析
		parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		// 驗證 claims
		claims, ok := parsedToken.Claims.(*Claims)
		assert.True(t, ok)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, version, claims.Version)
		assert.Equal(t, "ecom-api", claims.Issuer)
		assert.NotEmpty(t, claims.ID) // JTI
	})

	t.Run("不同用戶生成不同的 token", func(t *testing.T) {
		token1, err := jwtSvc.GenerateAccessToken("user-1", 0)
		assert.NoError(t, err)

		token2, err := jwtSvc.GenerateAccessToken("user-2", 0)
		assert.NoError(t, err)

		assert.NotEqual(t, token1, token2)
	})

	t.Run("每次生成的 JTI 都不同", func(t *testing.T) {
		token1, err := jwtSvc.GenerateAccessToken("user-1", 0)
		assert.NoError(t, err)

		token2, err := jwtSvc.GenerateAccessToken("user-1", 0)
		assert.NoError(t, err)

		// 解析兩個 token
		claims1 := &Claims{}
		_, _ = jwt.ParseWithClaims(token1, claims1, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		claims2 := &Claims{}
		_, _ = jwt.ParseWithClaims(token2, claims2, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		assert.NotEqual(t, claims1.ID, claims2.ID)
	})
}

func TestJWTSvc_GenerateRefreshToken(t *testing.T) {
	mockTokenRepo := new(repo.MockTokenRepository)
	jwtSvc := setupJWTService(mockTokenRepo)

	t.Run("成功生成 refresh token", func(t *testing.T) {
		userID := "user-456"
		version := 2

		token, err := jwtSvc.GenerateRefreshToken(userID, version)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// 驗證 token 可以被解析
		parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(*Claims)
		assert.True(t, ok)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, version, claims.Version)
	})
}

func TestJWTSvc_ValidateTokenBasic(t *testing.T) {
	mockTokenRepo := new(repo.MockTokenRepository)
	jwtSvc := setupJWTService(mockTokenRepo)

	t.Run("成功驗證有效的 token", func(t *testing.T) {
		userID := "user-789"
		version := 3

		token, err := jwtSvc.GenerateAccessToken(userID, version)
		require.NoError(t, err)

		tokenDetails, err := jwtSvc.ValidateTokenBasic(token)
		assert.NoError(t, err)
		assert.NotNil(t, tokenDetails)
		assert.Equal(t, userID, tokenDetails.UserID)
		assert.Equal(t, version, tokenDetails.Version)
		assert.NotEmpty(t, tokenDetails.JTI)
	})

	t.Run("驗證失敗 - 無效的 token", func(t *testing.T) {
		tokenDetails, err := jwtSvc.ValidateTokenBasic("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, tokenDetails)
	})

	t.Run("驗證失敗 - 過期的 token", func(t *testing.T) {
		// 創建一個過期的 token
		expiredSvc := &JWTSvc{
			secretKey:         "test-secret-key",
			accessTokenExpiry: -1 * time.Hour, // 負數，立即過期
			tokenRepo:         mockTokenRepo,
		}

		token, err := expiredSvc.GenerateAccessToken("user-1", 0)
		require.NoError(t, err)

		// 等待一小段時間確保過期
		time.Sleep(10 * time.Millisecond)

		tokenDetails, err := jwtSvc.ValidateTokenBasic(token)
		assert.Error(t, err)
		assert.Nil(t, tokenDetails)
	})

	t.Run("驗證失敗 - 錯誤的簽名", func(t *testing.T) {
		// 用不同的 secret 生成 token
		otherSvc := &JWTSvc{
			secretKey:         "different-secret",
			accessTokenExpiry: 15 * time.Minute,
			tokenRepo:         mockTokenRepo,
		}

		token, err := otherSvc.GenerateAccessToken("user-1", 0)
		require.NoError(t, err)

		// 用原來的 secret 驗證
		tokenDetails, err := jwtSvc.ValidateTokenBasic(token)
		assert.Error(t, err)
		assert.Nil(t, tokenDetails)
	})
}

func TestJWTSvc_ValidateToken(t *testing.T) {
	mockTokenRepo := new(repo.MockTokenRepository)
	jwtSvc := setupJWTService(mockTokenRepo)
	ctx := context.Background()

	t.Run("成功驗證（未被加入黑名單且版本正確）", func(t *testing.T) {
		userID := "user-valid"
		version := 5

		token, err := jwtSvc.GenerateAccessToken(userID, version)
		require.NoError(t, err)

		// 解析 token 獲取 JTI
		claims := &Claims{}
		_, _ = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		// 設置 mock 預期
		mockTokenRepo.On("IsBlacklisted", ctx, claims.ID).Return(false, nil).Once()
		mockTokenRepo.On("GetUserVersion", ctx, userID).Return(version, nil).Once()

		tokenDetails, err := jwtSvc.ValidateToken(ctx, token)
		assert.NoError(t, err)
		assert.NotNil(t, tokenDetails)
		assert.Equal(t, userID, tokenDetails.UserID)
		assert.Equal(t, version, tokenDetails.Version)

		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("驗證失敗 - token 在黑名單中", func(t *testing.T) {
		userID := "user-blacklisted"
		version := 1

		token, err := jwtSvc.GenerateAccessToken(userID, version)
		require.NoError(t, err)

		claims := &Claims{}
		_, _ = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		mockTokenRepo.On("IsBlacklisted", ctx, claims.ID).Return(true, nil).Once()

		tokenDetails, err := jwtSvc.ValidateToken(ctx, token)
		assert.Error(t, err)
		assert.Nil(t, tokenDetails)
		assert.Contains(t, err.Error(), "revoked")

		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("驗證失敗 - 版本不匹配", func(t *testing.T) {
		userID := "user-version-mismatch"
		tokenVersion := 1
		currentVersion := 5

		token, err := jwtSvc.GenerateAccessToken(userID, tokenVersion)
		require.NoError(t, err)

		claims := &Claims{}
		_, _ = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		mockTokenRepo.On("IsBlacklisted", ctx, claims.ID).Return(false, nil).Once()
		mockTokenRepo.On("GetUserVersion", ctx, userID).Return(currentVersion, nil).Once()

		tokenDetails, err := jwtSvc.ValidateToken(ctx, token)
		assert.Error(t, err)
		assert.Nil(t, tokenDetails)
		assert.Contains(t, err.Error(), "invalidated")

		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("驗證失敗 - Redis 錯誤（檢查黑名單）", func(t *testing.T) {
		userID := "user-redis-error"
		version := 1

		token, err := jwtSvc.GenerateAccessToken(userID, version)
		require.NoError(t, err)

		claims := &Claims{}
		_, _ = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		mockTokenRepo.On("IsBlacklisted", ctx, claims.ID).Return(false, errors.New("redis connection error")).Once()

		tokenDetails, err := jwtSvc.ValidateToken(ctx, token)
		assert.Error(t, err)
		assert.Nil(t, tokenDetails)
		assert.Contains(t, err.Error(), "failed to verify token")

		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("驗證失敗 - Redis 錯誤（獲取版本）", func(t *testing.T) {
		userID := "user-version-error"
		version := 1

		token, err := jwtSvc.GenerateAccessToken(userID, version)
		require.NoError(t, err)

		claims := &Claims{}
		_, _ = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		mockTokenRepo.On("IsBlacklisted", ctx, claims.ID).Return(false, nil).Once()
		mockTokenRepo.On("GetUserVersion", ctx, userID).Return(0, errors.New("redis error")).Once()

		tokenDetails, err := jwtSvc.ValidateToken(ctx, token)
		assert.Error(t, err)
		assert.Nil(t, tokenDetails)

		mockTokenRepo.AssertExpectations(t)
	})
}
