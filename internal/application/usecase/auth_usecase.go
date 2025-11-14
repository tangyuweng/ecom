package usecase

import (
	"context"
	"errors"

	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
)

type JWTService interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateToken(token string) (string, error)
}

type AuthUseCase struct {
	userRepo   repository.UserRepository
	jwtService JWTService
}

func NewAuthUseCase(userRepo repository.UserRepository, jwtService JWTService) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		jwtService: jwtService,
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
		return nil, errors.New("invalid email or password")
	}

	if !user.CheckPassword(req.Password) {
		return nil, errors.New("invalid email or password")
	}

	accessToken, err := uc.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.jwtService.GenerateRefreshToken(user.ID)
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
	userID, err := uc.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	_, err = uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, entity.ErrUserNotFound
	}

	newAccessToken, err := uc.jwtService.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{AccessToken: newAccessToken}, nil
}
