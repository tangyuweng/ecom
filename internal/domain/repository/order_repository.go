package repository

import (
	"context"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type OrderRepository interface {
	// Order operations
	Create(ctx context.Context, order *entity.Order) error
	FindByID(ctx context.Context, orderID string) (*entity.Order, error)
	FindByUserID(ctx context.Context, userID string) ([]*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error

	// OrderItem operations
	CreateItems(ctx context.Context, items []*entity.OrderItem) error
	FindItemsByOrderID(ctx context.Context, orderID string) ([]*entity.OrderItem, error)
}
