package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	repo "github.com/tangyuweng/ecom/internal/domain/repository"
	svc "github.com/tangyuweng/ecom/internal/domain/service"
)

func TestAuthUseCase_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("成功註冊新用戶", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		req := dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
			Phone:    "1234567890",
		}

		mockUserRepo.On("ExistsByEmail", ctx, req.Email).Return(false, nil).Once()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(nil).Once()
		mockTokenRepo.On("SetUserVersion", ctx, mock.AnythingOfType("string"), 0).Return(nil).Once()

		resp, err := authUseCase.Register(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Email, resp.User.Email)
		assert.Equal(t, req.Name, resp.User.Name)
		assert.Equal(t, req.Phone, resp.User.Phone)
		assert.Equal(t, "user", resp.User.Role)

		mockUserRepo.AssertExpectations(t)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("註冊失敗 - email 已存在", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		req := dto.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "Test User",
			Phone:    "1234567890",
		}

		// 模擬 email 已存在返回 true
		mockUserRepo.On("ExistsByEmail", ctx, req.Email).Return(true, nil).Once()

		resp, err := authUseCase.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "email already exists")

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("成功登入", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		user, _ := entity.NewUser("test@example.com", "password123", "Test User", "1234567890")
		user.ID = "user-123"

		req := dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		version := 0
		accessToken := "access-token-123"
		refreshToken := "refresh-token-456"

		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil).Once()
		mockTokenRepo.On("GetUserVersion", ctx, user.ID).Return(version, nil).Once()
		mockJWTSvc.On("GenerateAccessToken", user.ID, version).Return(accessToken, nil).Once()
		mockJWTSvc.On("GenerateRefreshToken", user.ID, version).Return(refreshToken, nil).Once()

		resp, err := authUseCase.Login(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, accessToken, resp.AccessToken)
		assert.Equal(t, refreshToken, resp.RefreshToken)
		assert.Equal(t, user.ID, resp.User.ID)
		assert.Equal(t, user.Email, resp.User.Email)

		mockUserRepo.AssertExpectations(t)
		mockTokenRepo.AssertExpectations(t)
		mockJWTSvc.AssertExpectations(t)
	})

	t.Run("登入失敗 - 用戶不存在", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		req := dto.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(nil, entity.ErrUserNotFound).Once()

		resp, err := authUseCase.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, entity.ErrInvalidCredentials, err)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("登入失敗 - 密碼錯誤", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		user, _ := entity.NewUser("test@example.com", "correctpassword", "Test User", "1234567890")

		req := dto.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		mockUserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil).Once()

		resp, err := authUseCase.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, entity.ErrInvalidCredentials, err)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_RefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("成功刷新 token", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		user, _ := entity.NewUser("test@example.com", "password123", "Test User", "1234567890")
		user.ID = "user-123"

		refreshToken := "valid-refresh-token"
		tokenDetails := &svc.TokenDetails{
			UserID:    user.ID,
			JTI:       "jti-123",
			Version:   0,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		newAccessToken := "new-access-token"

		mockJWTSvc.On("ValidateToken", ctx, refreshToken).Return(tokenDetails, nil).Once()
		mockUserRepo.On("FindByID", ctx, user.ID).Return(user, nil).Once()
		mockJWTSvc.On("GenerateAccessToken", user.ID, tokenDetails.Version).Return(newAccessToken, nil).Once()

		resp, err := authUseCase.RefreshToken(ctx, refreshToken)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, newAccessToken, resp.AccessToken)

		mockJWTSvc.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("刷新失敗 - refresh token 無效", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		refreshToken := "invalid-refresh-token"

		mockJWTSvc.On("ValidateToken", ctx, refreshToken).Return(nil, errors.New("invalid token")).Once()

		resp, err := authUseCase.RefreshToken(ctx, refreshToken)

		assert.Error(t, err)
		assert.Nil(t, resp)

		mockJWTSvc.AssertExpectations(t)
	})

	t.Run("刷新失敗 - token 在黑名單中", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		refreshToken := "blacklisted-token"

		mockJWTSvc.On("ValidateToken", ctx, refreshToken).Return(nil, errors.New("token has been revoked")).Once()

		resp, err := authUseCase.RefreshToken(ctx, refreshToken)

		assert.Error(t, err)
		assert.Nil(t, resp)

		mockJWTSvc.AssertExpectations(t)
	})

	t.Run("刷新失敗 - 用戶不存在", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		refreshToken := "valid-token-but-user-deleted"
		tokenDetails := &svc.TokenDetails{
			UserID:    "deleted-user",
			JTI:       "jti-123",
			Version:   0,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		mockJWTSvc.On("ValidateToken", ctx, refreshToken).Return(tokenDetails, nil).Once()
		mockUserRepo.On("FindByID", ctx, tokenDetails.UserID).Return(nil, entity.ErrUserNotFound).Once()

		resp, err := authUseCase.RefreshToken(ctx, refreshToken)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, entity.ErrUserNotFound, err)

		mockJWTSvc.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthUseCase_Logout(t *testing.T) {
	ctx := context.Background()

	t.Run("成功登出", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		refreshToken := "valid-refresh-token"
		tokenDetails := &svc.TokenDetails{
			UserID:    "user-123",
			JTI:       "jti-456",
			Version:   0,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		ttl := time.Until(tokenDetails.ExpiresAt)

		mockJWTSvc.On("ValidateTokenBasic", refreshToken).Return(tokenDetails, nil).Once()
		mockTokenRepo.On("AddToBlacklist", ctx, tokenDetails.JTI, mock.MatchedBy(func(d time.Duration) bool {
			// TTL 應該接近預期值（允許些微差異）
			return d > 0 && d <= ttl
		})).Return(nil).Once()

		err := authUseCase.Logout(ctx, refreshToken)

		assert.NoError(t, err)

		mockJWTSvc.AssertExpectations(t)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("登出失敗 - token 無效（不拋出錯誤）", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		refreshToken := "invalid-token"

		mockJWTSvc.On("ValidateTokenBasic", refreshToken).Return(nil, errors.New("invalid token")).Once()

		err := authUseCase.Logout(ctx, refreshToken)

		// 即使 token 無效也不應該報錯
		assert.NoError(t, err)

		mockJWTSvc.AssertExpectations(t)
	})

	t.Run("登出時 token 已過期（不加入黑名單）", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		refreshToken := "expired-token"
		tokenDetails := &svc.TokenDetails{
			UserID:    "user-123",
			JTI:       "jti-789",
			Version:   0,
			ExpiresAt: time.Now().Add(-1 * time.Hour), // 已過期
		}

		mockJWTSvc.On("ValidateTokenBasic", refreshToken).Return(tokenDetails, nil).Once()

		err := authUseCase.Logout(ctx, refreshToken)

		assert.NoError(t, err)

		mockJWTSvc.AssertExpectations(t)
		// 不應該調用 AddToBlacklist
		mockTokenRepo.AssertNotCalled(t, "AddToBlacklist")
	})
}

func TestAuthUseCase_LogoutAll(t *testing.T) {
	ctx := context.Background()

	t.Run("成功撤銷用戶所有 token", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		userID := "user-123"

		mockTokenRepo.On("IncrementUserVersion", ctx, userID).Return(nil).Once()

		err := authUseCase.LogoutAll(ctx, userID)

		assert.NoError(t, err)

		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("撤銷失敗 - Redis 錯誤", func(t *testing.T) {
		mockUserRepo := new(repo.MockUserRepository)
		mockTokenRepo := new(repo.MockTokenRepository)
		mockJWTSvc := new(svc.MockJWTService)

		authUseCase := NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTSvc)

		userID := "user-456"

		mockTokenRepo.On("IncrementUserVersion", ctx, userID).Return(errors.New("redis error")).Once()

		err := authUseCase.LogoutAll(ctx, userID)

		assert.Error(t, err)

		mockTokenRepo.AssertExpectations(t)
	})
}
