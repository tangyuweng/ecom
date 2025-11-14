package repository

import (
	"context"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindAll(ctx context.Context) ([]*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
