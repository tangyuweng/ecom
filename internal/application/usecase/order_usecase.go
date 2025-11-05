package usecase

import (
	"context"

	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
)

type OrderUseCase struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewOrderUseCase(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (uc *OrderUseCase) CreateOrderFromCart(ctx context.Context, userID string, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	cart, err := uc.cartRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if cart.IsEmpty() {
		return nil, entity.ErrEmptyCart
	}

	// 3. 驗證所有商品的庫存並準備訂單項目
	orderItems := make([]*entity.OrderItem, 0, len(cart.Items))
	for _, cartItem := range cart.Items {
		// 取得最新的商品資訊
		product, err := uc.productRepo.FindByID(ctx, cartItem.ProductID)
		if err != nil {
			return nil, err
		}

		// 檢查商品是否啟用
		if !product.IsActive {
			return nil, entity.ErrProductNotFound
		}

		// 檢查庫存是否足夠
		if product.StockQuantity < cartItem.Quantity {
			return nil, entity.ErrProductInsufficientStock
		}

		// 創建訂單項目（使用當前價格）
		orderItem, err := entity.NewOrderItem("", cartItem.ProductID, cartItem.Quantity, product.Price)
		if err != nil {
			return nil, err
		}
		orderItem.Product = product
		orderItems = append(orderItems, orderItem)
	}

	// 4. 創建訂單
	order, err := entity.NewOrder(userID, req.ShippingAddress, req.RecipientName, orderItems)
	if err != nil {
		return nil, err
	}

	// 5. 儲存訂單（在 transaction 中處理）
	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 6. 扣除商品庫存
	for _, item := range orderItems {
		product, _ := uc.productRepo.FindByID(ctx, item.ProductID)
		newStock := product.StockQuantity - item.Quantity
		if err := product.Update(
			product.CategoryID,
			product.Name,
			product.Description,
			product.Price,
			newStock,
			product.IsActive,
		); err != nil {
			return nil, err
		}
		if err := uc.productRepo.Update(ctx, product); err != nil {
			return nil, err
		}
	}

	// 7. 清空購物車
	if err := uc.cartRepo.Delete(ctx, userID); err != nil {
		return nil, err
	}

	// 8. 返回訂單資訊
	createdOrder, err := uc.orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	return uc.toOrderResponse(createdOrder), nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, userID, orderID string) (*dto.OrderResponse, error) {
	order, err := uc.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, entity.ErrOrderUnauthorized
	}

	return uc.toOrderResponse(order), nil
}

func (uc *OrderUseCase) GetUserOrders(ctx context.Context, userID string) ([]*dto.OrderResponse, error) {
	orders, err := uc.orderRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = uc.toOrderResponse(order)
	}

	return responses, nil
}

// CancelOrder 取消訂單（用戶可用）
func (uc *OrderUseCase) CancelOrder(ctx context.Context, userID, orderID string) (*dto.OrderResponse, error) {
	order, err := uc.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, entity.ErrOrderUnauthorized
	}

	// 取消訂單
	if err := order.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	// 恢復商品庫存
	for _, item := range order.Items {
		product, err := uc.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			continue // 如果商品已被刪除，跳過
		}

		newStock := product.StockQuantity + item.Quantity
		if err := product.Update(
			product.CategoryID,
			product.Name,
			product.Description,
			product.Price,
			newStock,
			product.IsActive,
		); err != nil {
			continue
		}
		uc.productRepo.Update(ctx, product)
	}

	updatedOrder, err := uc.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return uc.toOrderResponse(updatedOrder), nil
}

func (uc *OrderUseCase) toOrderResponse(order *entity.Order) *dto.OrderResponse {
	items := make([]*dto.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = uc.toOrderItemResponse(item)
	}

	return &dto.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		OrderDate:       order.OrderDate,
		TotalAmount:     order.TotalAmount,
		Status:          string(order.Status),
		ShippingAddress: order.ShippingAddress,
		RecipientName:   order.RecipientName,
		Items:           items,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

func (uc *OrderUseCase) toOrderItemResponse(item *entity.OrderItem) *dto.OrderItemResponse {
	response := &dto.OrderItemResponse{
		ID:        item.ID,
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
		UnitPrice: item.UnitPrice,
		Subtotal:  item.Subtotal,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	if item.Product != nil {
		response.ProductName = item.Product.Name
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
