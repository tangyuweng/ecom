package repository

import (
	"context"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.Product, error)
	FindByCategoryID(ctx context.Context, id string) ([]*entity.Product, error)
	FindByQuery(ctx context.Context, query *entity.ProductQuery) ([]*entity.Product, int, error)
	UpdateStock(ctx context.Context, id string, quantity int) error
	ExistsByCategoryID(ctx context.Context, categoryID string) (bool, error)
}
