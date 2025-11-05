package repository

import (
	"context"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type CartRepository interface {
	// Cart operations
	FindByUserID(ctx context.Context, userID string) (*entity.Cart, error)
	Create(ctx context.Context, cart *entity.Cart) error
	Delete(ctx context.Context, userID string) error

	// CartItem operations
	AddItem(ctx context.Context, item *entity.CartItem) error
	UpdateItem(ctx context.Context, item *entity.CartItem) error
	RemoveItem(ctx context.Context, cartItemID string) error
	FindItemByID(ctx context.Context, cartItemID string) (*entity.CartItem, error)
	FindItemByCartAndProduct(ctx context.Context, cartID, productID string) (*entity.CartItem, error)
}
