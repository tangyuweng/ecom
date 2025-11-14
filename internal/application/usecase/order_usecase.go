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
	userRepo    repository.UserRepository
}

func NewOrderUseCase(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	userRepo repository.UserRepository,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
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

	// 驗證所有商品的庫存並準備訂單項目
	orderItems := make([]*entity.OrderItem, 0, len(cart.Items))
	for _, cartItem := range cart.Items {
		product, err := uc.productRepo.FindByID(ctx, cartItem.ProductID)
		if err != nil {
			return nil, err
		}

		if !product.IsActive {
			return nil, entity.ErrProductNotFound
		}

		if product.StockQuantity < cartItem.Quantity {
			return nil, entity.ErrProductInsufficientStock
		}

		// 使用當前價格創建訂單項目
		orderItem, err := entity.NewOrderItem("", cartItem.ProductID, cartItem.Quantity, product.Price)
		if err != nil {
			return nil, err
		}
		orderItem.Product = product
		orderItems = append(orderItems, orderItem)
	}

	order, err := entity.NewOrder(userID, req.ShippingAddress, req.RecipientName, orderItems)
	if err != nil {
		return nil, err
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 扣除商品庫存
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

	// 清空購物車
	if err := uc.cartRepo.Delete(ctx, userID); err != nil {
		return nil, err
	}

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

// 取消訂單（用戶可用，僅限 Pending 或 Processing 狀態）
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

// 更新訂單狀態 (僅限管理員，若是以經是 Completed 或 Cancelled 狀態就不能操作了，若取消訂單則恢復庫存)
func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, adminID, orderID string, req dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error) {
	admin, err := uc.userRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	if !admin.IsAdmin() {
		return nil, entity.ErrUserUnauthorized
	}

	order, err := uc.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 檢查訂單當前狀態是否允許更新
	if order.Status == entity.OrderStatusCompleted || order.Status == entity.OrderStatusCancelled {
		return nil, entity.ErrOrderInvalidStatusTransition
	}

	newStatus := entity.OrderStatus(req.Status)

	// 根據新狀態決定操作
	if newStatus == entity.OrderStatusCancelled {
		if err := order.Cancel(); err != nil {
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
	} else {
		// 其他狀態轉換使用 UpdateStatus 方法
		if err := order.UpdateStatus(newStatus); err != nil {
			return nil, err
		}
	}

	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
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
