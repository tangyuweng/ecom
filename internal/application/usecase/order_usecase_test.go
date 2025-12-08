package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tangyuweng/ecom/internal/application/dto"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	repo "github.com/tangyuweng/ecom/internal/domain/repository"
	svc "github.com/tangyuweng/ecom/internal/domain/service"
)

// ===== Test Cases =====

// TestCreateOrderFromCart_Success 測試成功創建訂單的情境
func TestCreateOrderFromCart_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"

	// 準備測試數據
	product, _ := entity.NewProduct("cat-1", "Test Product", "Description", 100, 10, true)
	product.ID = "prod-1"

	cartItem, _ := entity.NewCartItem("cart-1", "prod-1", 2)
	cart := entity.NewCart(userID)
	cart.ID = "cart-1"
	cart.Items = []*entity.CartItem{cartItem}

	orderItem, _ := entity.NewOrderItem("", "prod-1", 2, 100)
	orderItem.Product = product
	order, _ := entity.NewOrder(userID, "Address", "Name", []*entity.OrderItem{orderItem})
	order.ID = "order-123"

	// 創建 mocks
	mockTxManager := new(repo.MockTransactionManager)
	mockOrderRepo := new(repo.MockOrderRepository)
	mockCartRepo := new(repo.MockCartRepository)
	mockProductRepo := new(repo.MockProductRepository)
	mockUserRepo := new(repo.MockUserRepository)
	mockNotificationSvc := new(svc.MockNotificationService)

	// 設定期望行為
	mockTxManager.On("WithTransaction", ctx).Return(nil)
	mockCartRepo.On("FindByUserID", mock.Anything, userID).Return(cart, nil)
	mockProductRepo.On("FindByID", mock.Anything, "prod-1").Return(product, nil)
	mockOrderRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Order")).Return(nil)
	mockProductRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(nil)
	mockCartRepo.On("Delete", mock.Anything, userID).Return(nil)
	mockOrderRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(order, nil)
	mockUserRepo.On("FindByAdmin", ctx).Return([]*entity.User{}, nil)

	// 創建 usecase
	uc := NewOrderUseCase(
		mockOrderRepo,
		mockCartRepo,
		mockProductRepo,
		mockUserRepo,
		mockNotificationSvc,
		mockTxManager,
	)

	req := dto.CreateOrderRequest{
		ShippingAddress: "Address",
		RecipientName:   "Name",
	}

	// Act
	result, err := uc.CreateOrderFromCart(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "order-123", result.ID)

	// 驗證所有 mock 都被正確調用
	mockTxManager.AssertExpectations(t)
	mockCartRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
}

// TestCreateOrderFromCart_EmptyCart 測試購物車為空的情境
func TestCreateOrderFromCart_EmptyCart(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"

	emptyCart := entity.NewCart(userID)

	mockTxManager := new(repo.MockTransactionManager)
	mockOrderRepo := new(repo.MockOrderRepository)
	mockCartRepo := new(repo.MockCartRepository)
	mockProductRepo := new(repo.MockProductRepository)
	mockUserRepo := new(repo.MockUserRepository)
	mockNotificationSvc := new(svc.MockNotificationService)

	// 事務會執行但返回錯誤
	mockTxManager.On("WithTransaction", ctx).Return(nil)
	mockCartRepo.On("FindByUserID", mock.Anything, userID).Return(emptyCart, nil)

	uc := NewOrderUseCase(
		mockOrderRepo,
		mockCartRepo,
		mockProductRepo,
		mockUserRepo,
		mockNotificationSvc,
		mockTxManager,
	)

	req := dto.CreateOrderRequest{
		ShippingAddress: "Address",
		RecipientName:   "Name",
	}

	// Act
	result, err := uc.CreateOrderFromCart(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, entity.ErrEmptyCart, err)

	mockCartRepo.AssertExpectations(t)
}

// TestCreateOrderFromCart_InsufficientStock 測試庫存不足的情境（應該回滾事務）
func TestCreateOrderFromCart_InsufficientStock(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"

	// 庫存只有 1，但要購買 2
	product, _ := entity.NewProduct("cat-1", "Test Product", "Description", 100, 1, true)
	product.ID = "prod-1"

	cartItem, _ := entity.NewCartItem("cart-1", "prod-1", 2)
	cart := entity.NewCart(userID)
	cart.ID = "cart-1"
	cart.Items = []*entity.CartItem{cartItem}

	mockTxManager := new(repo.MockTransactionManager)
	mockOrderRepo := new(repo.MockOrderRepository)
	mockCartRepo := new(repo.MockCartRepository)
	mockProductRepo := new(repo.MockProductRepository)
	mockUserRepo := new(repo.MockUserRepository)
	mockNotificationSvc := new(svc.MockNotificationService)

	// 事務會執行但應該回滾
	mockTxManager.On("WithTransaction", ctx).Return(nil)
	mockCartRepo.On("FindByUserID", mock.Anything, userID).Return(cart, nil)
	mockProductRepo.On("FindByID", mock.Anything, "prod-1").Return(product, nil)

	uc := NewOrderUseCase(
		mockOrderRepo,
		mockCartRepo,
		mockProductRepo,
		mockUserRepo,
		mockNotificationSvc,
		mockTxManager,
	)

	req := dto.CreateOrderRequest{
		ShippingAddress: "Address",
		RecipientName:   "Name",
	}

	// Act
	result, err := uc.CreateOrderFromCart(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, entity.ErrProductInsufficientStock, err)

	// 確認沒有調用 Create（因為事務應該在檢查庫存時就失敗了）
	mockOrderRepo.AssertNotCalled(t, "Create")
}

// TestCancelOrder_Success 測試成功取消訂單並恢復庫存
func TestCancelOrder_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"
	orderID := "order-123"

	product, _ := entity.NewProduct("cat-1", "Test Product", "Description", 100, 8, true)
	product.ID = "prod-1"

	orderItem, _ := entity.NewOrderItem("", "prod-1", 2, 100)
	order, _ := entity.NewOrder(userID, "Address", "Name", []*entity.OrderItem{orderItem})
	order.ID = orderID
	order.Items = []*entity.OrderItem{orderItem}

	mockTxManager := new(repo.MockTransactionManager)
	mockOrderRepo := new(repo.MockOrderRepository)
	mockCartRepo := new(repo.MockCartRepository)
	mockProductRepo := new(repo.MockProductRepository)
	mockUserRepo := new(repo.MockUserRepository)
	mockNotificationSvc := new(svc.MockNotificationService)

	mockTxManager.On("WithTransaction", ctx).Return(nil)
	mockOrderRepo.On("FindByID", mock.Anything, orderID).Return(order, nil)
	mockOrderRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Order")).Return(nil)
	mockProductRepo.On("FindByID", mock.Anything, "prod-1").Return(product, nil)
	mockProductRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(nil)
	mockUserRepo.On("FindByAdmin", ctx).Return([]*entity.User{}, nil)
	mockNotificationSvc.On("NotifyUser", ctx, userID, mock.AnythingOfType("*entity.Notification")).Return(nil)

	uc := NewOrderUseCase(
		mockOrderRepo,
		mockCartRepo,
		mockProductRepo,
		mockUserRepo,
		mockNotificationSvc,
		mockTxManager,
	)

	// Act
	result, err := uc.CancelOrder(ctx, userID, orderID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(entity.OrderStatusCancelled), result.Status)

	// 驗證庫存恢復被調用
	mockProductRepo.AssertCalled(t, "Update", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
		// 庫存應該從 8 恢復到 10（8 + 2）
		return p.StockQuantity == 10
	}))
}

// TestCancelOrder_Unauthorized 測試未授權取消訂單
func TestCancelOrder_Unauthorized(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"
	otherUserID := "user-456"
	orderID := "order-123"

	orderItem, _ := entity.NewOrderItem("", "prod-1", 2, 100)
	order, _ := entity.NewOrder(otherUserID, "Address", "Name", []*entity.OrderItem{orderItem})
	order.ID = orderID

	mockTxManager := new(repo.MockTransactionManager)
	mockOrderRepo := new(repo.MockOrderRepository)
	mockCartRepo := new(repo.MockCartRepository)
	mockProductRepo := new(repo.MockProductRepository)
	mockUserRepo := new(repo.MockUserRepository)
	mockNotificationSvc := new(svc.MockNotificationService)

	mockTxManager.On("WithTransaction", ctx).Return(nil)
	mockOrderRepo.On("FindByID", mock.Anything, orderID).Return(order, nil)

	uc := NewOrderUseCase(
		mockOrderRepo,
		mockCartRepo,
		mockProductRepo,
		mockUserRepo,
		mockNotificationSvc,
		mockTxManager,
	)

	// Act
	result, err := uc.CancelOrder(ctx, userID, orderID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, entity.ErrOrderUnauthorized, err)

	// 確認沒有更新訂單
	mockOrderRepo.AssertNotCalled(t, "Update")
}

// TestTransactionRollback 測試事務回滾機制
func TestTransactionRollback_OrderCreationFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"

	product, _ := entity.NewProduct("cat-1", "Test Product", "Description", 100, 10, true)
	product.ID = "prod-1"

	cartItem, _ := entity.NewCartItem("cart-1", "prod-1", 2)
	cart := entity.NewCart(userID)
	cart.ID = "cart-1"
	cart.Items = []*entity.CartItem{cartItem}

	mockTxManager := new(repo.MockTransactionManager)
	mockOrderRepo := new(repo.MockOrderRepository)
	mockCartRepo := new(repo.MockCartRepository)
	mockProductRepo := new(repo.MockProductRepository)
	mockUserRepo := new(repo.MockUserRepository)
	mockNotificationSvc := new(svc.MockNotificationService)

	// 模擬事務執行但訂單創建失敗
	mockTxManager.On("WithTransaction", ctx).Return(nil)
	mockCartRepo.On("FindByUserID", mock.Anything, userID).Return(cart, nil)
	mockProductRepo.On("FindByID", mock.Anything, "prod-1").Return(product, nil)
	mockOrderRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Order")).Return(errors.New("database error"))

	uc := NewOrderUseCase(
		mockOrderRepo,
		mockCartRepo,
		mockProductRepo,
		mockUserRepo,
		mockNotificationSvc,
		mockTxManager,
	)

	req := dto.CreateOrderRequest{
		ShippingAddress: "Address",
		RecipientName:   "Name",
	}

	// Act
	result, err := uc.CreateOrderFromCart(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")

	// 確認後續操作沒有被調用（因為事務應該回滾）
	mockProductRepo.AssertNotCalled(t, "Update")
	mockCartRepo.AssertNotCalled(t, "Delete")
}
