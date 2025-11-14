package usecase

import (
	"context"

	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
)

type UserUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(userRepo repository.UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

func (uc *UserUseCase) GetUsers(ctx context.Context, userID string) (*dto.UserListResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsAdmin() {
		return nil, entity.ErrUserUnauthorized
	}

	users, err := uc.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = uc.toUserResponse(user)
	}

	return &dto.UserListResponse{
		Total: len(userResponses),
		Users: userResponses,
	}, nil
}

// func (uc *UserUseCase) GetUser(ctx context.Context, userID string) (*dto.UserResponse, error) {
// 	user, err := uc.userRepo.FindByID(ctx, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if !user.IsAdmin() || user.ID != userID {
// 		return nil, entity.ErrUserUnauthorized
// 	}

// 	return uc.toUserResponse(user), nil
// }

func (uc *UserUseCase) UpdateRole(ctx context.Context, adminID, userID string, req dto.UpdateUserRoleRequest) (*dto.UserResponse, error) {
	admin, err := uc.userRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	if !admin.IsAdmin() {
		return nil, entity.ErrUserUnauthorized
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := user.UpdateRole(entity.UserRole(req.UserRole)); err != nil {
		return nil, err
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	updateUser, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return uc.toUserResponse(updateUser), nil
}

func (uc *UserUseCase) toUserResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Phone: user.Phone,
		Role:  string(user.Role),
	}
}
