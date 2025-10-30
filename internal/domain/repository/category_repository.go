package repository

import (
	"context"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	FindByID(ctx context.Context, id string) (*entity.Category, error)
	FindByName(ctx context.Context, name string) (*entity.Category, error)
	FindAll(ctx context.Context) ([]*entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id string) error
	ExistsByName(ctx context.Context, name string) (bool, error)
}
