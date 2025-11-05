package usecase

import (
	"context"

	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
)

type CartUseCase struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartUseCase(
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
) *CartUseCase {
	return &CartUseCase{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (uc *CartUseCase) GetCart(ctx context.Context, userID string) (*dto.CartResponse, error) {
	cart, err := uc.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		if err == entity.ErrCartNotFound {
			newCart := entity.NewCart(userID)
			if err := uc.cartRepo.Create(ctx, newCart); err != nil {
				return nil, err
			}
			return uc.toCartResponse(newCart), nil
		}
		return nil, err
	}

	return uc.toCartResponse(cart), nil
}

func (uc *CartUseCase) AddToCart(ctx context.Context, userID string, req dto.AddToCartRequest) (*dto.CartResponse, error) {
	product, err := uc.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	if !product.IsActive {
		return nil, entity.ErrProductNotFound
	}

	if product.StockQuantity < req.Quantity {
		return nil, entity.ErrProductInsufficientStock
	}

	cart, err := uc.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		if err == entity.ErrCartNotFound {
			cart = entity.NewCart(userID)
			if err := uc.cartRepo.Create(ctx, cart); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Check if product already exists in cart
	existingItem, err := uc.cartRepo.FindItemByCartAndProduct(ctx, cart.ID, req.ProductID)
	switch err {
	case nil:
		// Update existing item quantity
		newQuantity := existingItem.Quantity + req.Quantity

		if product.StockQuantity < newQuantity {
			return nil, entity.ErrProductInsufficientStock
		}

		if err := existingItem.UpdateQuantity(newQuantity); err != nil {
			return nil, err
		}
		if err := uc.cartRepo.UpdateItem(ctx, existingItem); err != nil {
			return nil, err
		}
	case entity.ErrCartItemNotFound:
		// Add new item to cart
		cartItem, err := entity.NewCartItem(cart.ID, req.ProductID, req.Quantity)
		if err != nil {
			return nil, err
		}

		if err := uc.cartRepo.AddItem(ctx, cartItem); err != nil {
			return nil, err
		}
	default:
		return nil, err
	}

	updatedCart, err := uc.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return uc.toCartResponse(updatedCart), nil
}

func (uc *CartUseCase) UpdateCartItem(ctx context.Context, userID, cartItemID string, req dto.UpdateCartItemRequest) (*dto.CartResponse, error) {
	cart, err := uc.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	cartItem, err := uc.cartRepo.FindItemByID(ctx, cartItemID)
	if err != nil {
		return nil, err
	}

	if cartItem.CartID != cart.ID {
		return nil, entity.ErrCartUnauthorized
	}

	// Verify product stock
	product, err := uc.productRepo.FindByID(ctx, cartItem.ProductID)
	if err != nil {
		return nil, err
	}

	if product.StockQuantity < req.Quantity {
		return nil, entity.ErrProductInsufficientStock
	}

	// Update quantity
	if err := cartItem.UpdateQuantity(req.Quantity); err != nil {
		return nil, err
	}

	if err := uc.cartRepo.UpdateItem(ctx, cartItem); err != nil {
		return nil, err
	}

	updatedCart, err := uc.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return uc.toCartResponse(updatedCart), nil
}

func (uc *CartUseCase) RemoveFromCart(ctx context.Context, userID, cartItemID string) error {
	cart, err := uc.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	cartItem, err := uc.cartRepo.FindItemByID(ctx, cartItemID)
	if err != nil {
		return err
	}

	if cartItem.CartID != cart.ID {
		return entity.ErrCartUnauthorized
	}

	return uc.cartRepo.RemoveItem(ctx, cartItemID)
}

func (uc *CartUseCase) ClearCart(ctx context.Context, userID string) error {
	_, err := uc.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return uc.cartRepo.Delete(ctx, userID)
}

func (uc *CartUseCase) toCartResponse(cart *entity.Cart) *dto.CartResponse {
	items := make([]*dto.CartItemResponse, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = uc.toCartItemResponse(item)
	}

	return &dto.CartResponse{
		ID:         cart.ID,
		UserID:     cart.UserID,
		Items:      items,
		TotalItems: cart.GetTotalItems(),
		TotalPrice: cart.GetTotalPrice(),
		CreatedAt:  cart.CreatedAt,
		UpdatedAt:  cart.UpdatedAt,
	}
}

func (uc *CartUseCase) toCartItemResponse(item *entity.CartItem) *dto.CartItemResponse {
	response := &dto.CartItemResponse{
		ID:        item.ID,
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	if item.Product != nil {
		response.ProductName = item.Product.Name
		response.ProductPrice = item.Product.Price
		response.Subtotal = item.Product.Price * float64(item.Quantity)
		response.Product = &dto.ProductResponse{
			ID:            item.Product.ID,
			Name:          item.Product.Name,
			Description:   item.Product.Description,
			Price:         item.Product.Price,
			StockQuantity: item.Product.StockQuantity,
			IsActive:      item.Product.IsActive,
			CategoryID:    item.Product.CategoryID,
			CategoryName:  item.Product.CategoryName,
			CategorySlug:  item.Product.CategorySlug,
			CreatedAt:     item.Product.CreatedAt,
			UpdatedAt:     item.Product.UpdatedAt,
		}
	}

	return response
}
