package repository

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/tangyuweng/ecom/internal/domain/entity"
)

// ===== Mock Implementations =====

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 直接執行函數，模擬事務的行為
	// 如果函數返回錯誤，事務會回滾（返回該錯誤）
	err := fn(ctx)

	// 調用 mock 記錄，但不使用其返回值來控制測試流程
	// 因為已經通過執行 fn(ctx) 來測試實際的業務邏輯
	m.Called(ctx)

	return err
}

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *entity.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *entity.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByQuery(ctx context.Context, query *entity.OrderQuery) ([]*entity.Order, int, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entity.Order), args.Int(1), args.Error(2)
}

func (m *MockOrderRepository) CreateItems(ctx context.Context, items []*entity.OrderItem) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *MockOrderRepository) FindItemsByOrderID(ctx context.Context, orderID string) ([]*entity.OrderItem, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.OrderItem), args.Error(1)
}

type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) FindByUserID(ctx context.Context, userID string) (*entity.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartRepository) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCartRepository) Create(ctx context.Context, cart *entity.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *MockCartRepository) AddItem(ctx context.Context, item *entity.CartItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCartRepository) UpdateItem(ctx context.Context, item *entity.CartItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCartRepository) RemoveItem(ctx context.Context, cartItemID string) error {
	args := m.Called(ctx, cartItemID)
	return args.Error(0)
}

func (m *MockCartRepository) FindItemByID(ctx context.Context, cartItemID string) (*entity.CartItem, error) {
	args := m.Called(ctx, cartItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CartItem), args.Error(1)
}

func (m *MockCartRepository) FindItemByCartAndProduct(ctx context.Context, cartID, productID string) (*entity.CartItem, error) {
	args := m.Called(ctx, cartID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CartItem), args.Error(1)
}

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Create(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) FindByCategoryID(ctx context.Context, categoryID string) ([]*entity.Product, error) {
	args := m.Called(ctx, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Product), args.Error(1)
}

func (m *MockProductRepository) FindByQuery(ctx context.Context, query *entity.ProductQuery) ([]*entity.Product, int, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entity.Product), args.Int(1), args.Error(2)
}

func (m *MockProductRepository) ExistsByCategoryID(ctx context.Context, categoryID string) (bool, error) {
	args := m.Called(ctx, categoryID)
	return args.Bool(0), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) FindAll(ctx context.Context) ([]*entity.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByAdmin(ctx context.Context) ([]*entity.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}
