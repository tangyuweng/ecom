package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
	"github.com/tangyuweng/ecom/internal/domain/service"
)

type AuthUseCase struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	jwtSvc    service.JWTService
}

func NewAuthUseCase(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, jwtSvc service.JWTService) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtSvc:    jwtSvc,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	exists, err := uc.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	user, err := entity.NewUser(req.Email, req.Password, req.Name, req.Phone)
	if err != nil {
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := uc.tokenRepo.SetUserVersion(ctx, user.ID, 0); err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Phone: user.Phone,
			Role:  string(user.Role),
		},
	}, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, entity.ErrInvalidCredentials
	}

	if !user.CheckPassword(req.Password) {
		return nil, entity.ErrInvalidCredentials
	}

	version, err := uc.tokenRepo.GetUserVersion(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	accessToken, err := uc.jwtSvc.GenerateAccessToken(user.ID, version)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtSvc.GenerateRefreshToken(user.ID, version)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Phone: user.Phone,
			Role:  string(user.Role),
		},
	}, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	tokenDetails, err := uc.jwtSvc.ValidateToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	_, err = uc.userRepo.FindByID(ctx, tokenDetails.UserID)
	if err != nil {
		return nil, entity.ErrUserNotFound
	}

	newAccessToken, err := uc.jwtSvc.GenerateAccessToken(tokenDetails.UserID, tokenDetails.Version)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{AccessToken: newAccessToken}, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	tokenDetails, err := uc.jwtSvc.ValidateTokenBasic(refreshToken)
	if err != nil {
		return nil
	}

	ttl := time.Until(tokenDetails.ExpiresAt)
	if ttl <= 0 {
		return nil // token 已過期，無須加入黑名單
	}

	return uc.tokenRepo.AddToBlacklist(ctx, tokenDetails.JTI, ttl)
}

func (uc *AuthUseCase) LogoutAll(ctx context.Context, userID string) error {
	return uc.tokenRepo.IncrementUserVersion(ctx, userID)
}
