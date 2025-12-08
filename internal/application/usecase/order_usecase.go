package usecase

import (
	"context"

	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
	"github.com/tangyuweng/ecom/internal/domain/service"
)

type OrderUseCase struct {
	orderRepo       repository.OrderRepository
	cartRepo        repository.CartRepository
	productRepo     repository.ProductRepository
	userRepo        repository.UserRepository
	notificationSvc service.NotificationService
	txManager       repository.TransactionManager
}

func NewOrderUseCase(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	userRepo repository.UserRepository,
	notificationSvc service.NotificationService,
	txManager repository.TransactionManager,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:       orderRepo,
		cartRepo:        cartRepo,
		productRepo:     productRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
		txManager:       txManager,
	}
}

func (uc *OrderUseCase) CreateOrderFromCart(ctx context.Context, userID string, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	var createdOrder *entity.Order

	// 使用事務包裹所有資料庫操作
	err := uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. 查詢購物車
		cart, err := uc.cartRepo.FindByUserID(txCtx, userID)
		if err != nil {
			return err
		}

		if cart.IsEmpty() {
			return entity.ErrEmptyCart
		}

		// 2. 驗證所有商品的庫存並準備訂單項目
		orderItems := make([]*entity.OrderItem, 0, len(cart.Items))
		for _, cartItem := range cart.Items {
			product, err := uc.productRepo.FindByID(txCtx, cartItem.ProductID)
			if err != nil {
				return err
			}

			if !product.IsActive {
				return entity.ErrProductNotFound
			}

			if product.StockQuantity < cartItem.Quantity {
				return entity.ErrProductInsufficientStock
			}

			// 使用當前價格創建訂單項目
			orderItem, err := entity.NewOrderItem("", cartItem.ProductID, cartItem.Quantity, product.Price)
			if err != nil {
				return err
			}
			orderItem.Product = product
			orderItems = append(orderItems, orderItem)
		}

		// 3. 創建訂單
		order, err := entity.NewOrder(userID, req.ShippingAddress, req.RecipientName, orderItems)
		if err != nil {
			return err
		}

		if err := uc.orderRepo.Create(txCtx, order); err != nil {
			return err
		}

		// 4. 扣除商品庫存
		for _, item := range orderItems {
			product, err := uc.productRepo.FindByID(txCtx, item.ProductID)
			if err != nil {
				return err
			}

			newStock := product.StockQuantity - item.Quantity
			if err := product.Update(
				product.CategoryID,
				product.Name,
				product.Description,
				product.Price,
				newStock,
				product.IsActive,
			); err != nil {
				return err
			}

			if err := uc.productRepo.Update(txCtx, product); err != nil {
				return err
			}
		}

		// 5. 清空購物車
		if err := uc.cartRepo.Delete(txCtx, userID); err != nil {
			return err
		}

		// 6. 查詢完整的訂單資料（包含關聯）
		createdOrder, err = uc.orderRepo.FindByID(txCtx, order.ID)
		if err != nil {
			return err
		}

		// 事務成功，所有操作都會 COMMIT
		return nil
	})

	if err != nil {
		// 事務失敗，所有操作都已 ROLLBACK
		return nil, err
	}

	notification := entity.NewNotification(
		entity.NotificationTypeNewOrderCreated,
		map[string]interface{}{
			"orderID":     createdOrder.ID,
			"userID":      createdOrder.UserID,
			"totalAmount": createdOrder.TotalAmount,
			"status":      string(createdOrder.Status),
			"message":     "新訂單已建立",
		},
	)

	admins, err := uc.userRepo.FindByAdmin(ctx)
	if err == nil {
		for _, admin := range admins {
			_ = uc.notificationSvc.NotifyUser(ctx, admin.ID, notification)
		}
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
	var updatedOrder *entity.Order

	err := uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		order, err := uc.orderRepo.FindByID(txCtx, orderID)
		if err != nil {
			return err
		}

		if order.UserID != userID {
			return entity.ErrOrderUnauthorized
		}

		// 取消訂單
		if err := order.Cancel(); err != nil {
			return err
		}

		if err := uc.orderRepo.Update(txCtx, order); err != nil {
			return err
		}

		// 恢復商品庫存
		for _, item := range order.Items {
			product, err := uc.productRepo.FindByID(txCtx, item.ProductID)
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

			if err := uc.productRepo.Update(txCtx, product); err != nil {
				return err
			}
		}

		updatedOrder, err = uc.orderRepo.FindByID(txCtx, orderID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	notification := entity.NewNotification(
		entity.NotificationTypeOrderStatusUpdated,
		map[string]interface{}{
			"orderID": updatedOrder.ID,
			"status":  string(updatedOrder.Status),
			"message": getStatusMessage(updatedOrder.Status),
		},
	)

	admins, err := uc.userRepo.FindByAdmin(ctx)
	if err == nil {
		for _, admin := range admins {
			_ = uc.notificationSvc.NotifyUser(ctx, admin.ID, notification)
		}
	}

	_ = uc.notificationSvc.NotifyUser(ctx, userID, notification)

	return uc.toOrderResponse(updatedOrder), nil
}

// 更新訂單狀態 (僅限管理員，若是以經是 Completed 或 Cancelled 狀態就不能操作了，若取消訂單則恢復庫存)
func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, adminID, orderID string, req dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error) {
	var updatedOrder *entity.Order

	err := uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		admin, err := uc.userRepo.FindByID(txCtx, adminID)
		if err != nil {
			return err
		}

		if !admin.IsAdmin() {
			return entity.ErrUserUnauthorized
		}

		order, err := uc.orderRepo.FindByID(txCtx, orderID)
		if err != nil {
			return err
		}

		// 檢查訂單當前狀態是否允許更新
		if order.Status == entity.OrderStatusCompleted || order.Status == entity.OrderStatusCancelled {
			return entity.ErrOrderInvalidStatusTransition
		}

		newStatus := entity.OrderStatus(req.Status)

		// 根據新狀態決定操作
		if newStatus == entity.OrderStatusCancelled {
			if err := order.Cancel(); err != nil {
				return err
			}

			// 恢復商品庫存
			for _, item := range order.Items {
				product, err := uc.productRepo.FindByID(txCtx, item.ProductID)
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
				if err := uc.productRepo.Update(txCtx, product); err != nil {
					return err
				}
			}
		} else {
			// 其他狀態轉換使用 UpdateStatus 方法
			if err := order.UpdateStatus(newStatus); err != nil {
				return err
			}
		}

		if err := uc.orderRepo.Update(txCtx, order); err != nil {
			return err
		}

		updatedOrder, err = uc.orderRepo.FindByID(txCtx, orderID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	notification := entity.NewNotification(
		entity.NotificationTypeOrderStatusUpdated,
		map[string]interface{}{
			"orderID": updatedOrder.ID,
			"status":  string(updatedOrder.Status),
			"message": getStatusMessage(updatedOrder.Status),
		},
	)
	_ = uc.notificationSvc.NotifyUser(ctx, updatedOrder.UserID, notification)

	return uc.toOrderResponse(updatedOrder), nil
}

// GetAllOrders 管理員獲取所有訂單（支援複合式查詢、分頁、排序）
func (uc *OrderUseCase) GetAllOrders(ctx context.Context, adminID string, req dto.OrderListRequest) (*dto.OrderListResponse, error) {
	admin, err := uc.userRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	if !admin.IsAdmin() {
		return nil, entity.ErrUserUnauthorized
	}

	// 設置默認值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	query := &entity.OrderQuery{
		UserID:    req.UserID,
		MinTotal:  req.MinTotal,
		MaxTotal:  req.MaxTotal,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}

	// 處理狀態過濾
	if req.Status != nil && *req.Status != "" {
		status := entity.OrderStatus(*req.Status)
		query.Status = &status
	}

	orders, total, err := uc.orderRepo.FindByQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	orderResponses := make([]*dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = uc.toOrderResponse(order)
	}

	return &dto.OrderListResponse{
		Total:  total,
		Orders: orderResponses,
	}, nil
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

// getStatusMessage 根據訂單狀態返回中文訊息
func getStatusMessage(status entity.OrderStatus) string {
	switch status {
	case entity.OrderStatusPending:
		return "訂單待處理"
	case entity.OrderStatusProcessing:
		return "訂單處理中"
	case entity.OrderStatusShipped:
		return "訂單已出貨"
	case entity.OrderStatusCompleted:
		return "訂單已完成"
	case entity.OrderStatusCancelled:
		return "訂單已取消"
	default:
		return "訂單狀態已更新"
	}
}
