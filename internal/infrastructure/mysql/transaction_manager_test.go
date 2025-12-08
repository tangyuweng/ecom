package mysql

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupTestDB 設置測試資料庫
// 注意：這需要一個真實的測試資料庫連接
func setupTestDB(t *testing.T) *gorm.DB {
	// 這裡使用測試資料庫連接
	// 實際使用時應該從環境變數讀取測試資料庫配置
	dsn := "root:password@tcp(localhost:3306)/ecom_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Skipping integration test: database not available")
		return nil
	}

	// 自動遷移測試表
	err = db.AutoMigrate(
		&entity.User{},
		&entity.Product{},
		&entity.Category{},
		&entity.Order{},
		&entity.OrderItem{},
	)
	require.NoError(t, err)

	return db
}

// cleanupTestDB 清理測試資料
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("DELETE FROM order_items")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM categories")
	db.Exec("DELETE FROM users")
}

// TestWithTransaction_Success 測試事務成功提交
func TestWithTransaction_Success(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	tm := NewGormTransactionManager(db)
	ctx := context.Background()

	// 準備測試數據
	user := &entity.User{
		ID:       "user-1",
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword123",
		Phone:    "1234567890",
		Role:     entity.RoleUser,
	}

	// Act - 在事務中創建用戶
	err := tm.WithTransaction(ctx, func(txCtx context.Context) error {
		txDB := GetDB(txCtx, db)
		return txDB.Create(user).Error
	})

	// Assert
	assert.NoError(t, err)

	// 驗證數據已提交到資料庫
	var foundUser entity.User
	err = db.First(&foundUser, "id = ?", "user-1").Error
	assert.NoError(t, err)
	assert.Equal(t, "testuser", foundUser.Name)
}

// TestWithTransaction_Rollback 測試事務回滾
func TestWithTransaction_Rollback(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	tm := NewGormTransactionManager(db)
	ctx := context.Background()

	user := &entity.User{
		ID:       "user-2",
		Name:     "testuser2",
		Email:    "test2@example.com",
		Password: "hashedpassword123",
		Phone:    "1234567890",
		Role:     entity.RoleUser,
	}

	expectedErr := errors.New("intentional error")

	// Act - 在事務中創建用戶後返回錯誤
	err := tm.WithTransaction(ctx, func(txCtx context.Context) error {
		txDB := GetDB(txCtx, db)
		if err := txDB.Create(user).Error; err != nil {
			return err
		}
		// 返回錯誤以觸發回滾
		return expectedErr
	})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	// 驗證數據已回滾，資料庫中不應該有這個用戶
	var foundUser entity.User
	err = db.First(&foundUser, "id = ?", "user-2").Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

// TestWithTransaction_PanicRecovery 測試 panic 處理
func TestWithTransaction_PanicRecovery(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	tm := NewGormTransactionManager(db)
	ctx := context.Background()

	user := &entity.User{
		ID:       "user-3",
		Name:     "testuser3",
		Email:    "test3@example.com",
		Password: "hashedpassword123",
		Phone:    "1234567890",
		Role:     entity.RoleUser,
	}

	// Act & Assert - 測試 panic 會被捕獲並重新拋出
	assert.Panics(t, func() {
		_ = tm.WithTransaction(ctx, func(txCtx context.Context) error {
			txDB := GetDB(txCtx, db)
			txDB.Create(user)
			panic("intentional panic")
		})
	})

	// 驗證數據已回滾
	var foundUser entity.User
	err := db.First(&foundUser, "id = ?", "user-3").Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

// TestWithTransaction_NestedOperations 測試事務中的多個操作
func TestWithTransaction_NestedOperations(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	tm := NewGormTransactionManager(db)
	ctx := context.Background()

	// 準備測試數據
	category := &entity.Category{
		ID:   "cat-1",
		Name: "Test Category",
		Slug: "test-category",
	}

	product := &entity.Product{
		ID:            "prod-1",
		CategoryID:    "cat-1",
		Name:          "Test Product",
		Description:   "Description",
		Price:         100,
		StockQuantity: 10,
		IsActive:      true,
	}

	// Act - 在同一事務中創建分類和產品
	err := tm.WithTransaction(ctx, func(txCtx context.Context) error {
		txDB := GetDB(txCtx, db)

		if err := txDB.Create(category).Error; err != nil {
			return err
		}

		if err := txDB.Create(product).Error; err != nil {
			return err
		}

		return nil
	})

	// Assert
	assert.NoError(t, err)

	// 驗證兩個實體都已創建
	var foundCategory entity.Category
	err = db.First(&foundCategory, "id = ?", "cat-1").Error
	assert.NoError(t, err)

	var foundProduct entity.Product
	err = db.First(&foundProduct, "id = ?", "prod-1").Error
	assert.NoError(t, err)
}

// TestWithTransaction_PartialRollback 測試部分操作失敗時的完整回滾
func TestWithTransaction_PartialRollback(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanupTestDB(t, db)

	tm := NewGormTransactionManager(db)
	ctx := context.Background()

	category := &entity.Category{
		ID:   "cat-2",
		Name: "Test Category 2",
		Slug: "test-category-2",
	}

	product := &entity.Product{
		ID:            "prod-2",
		CategoryID:    "cat-2",
		Name:          "Test Product 2",
		Description:   "Description",
		Price:         100,
		StockQuantity: 10,
		IsActive:      true,
	}

	// Act - 創建分類成功，但產品失敗
	err := tm.WithTransaction(ctx, func(txCtx context.Context) error {
		txDB := GetDB(txCtx, db)

		// 第一個操作成功
		if err := txDB.Create(category).Error; err != nil {
			return err
		}

		// 第二個操作成功
		if err := txDB.Create(product).Error; err != nil {
			return err
		}

		// 但最後返回錯誤
		return errors.New("rollback all")
	})

	// Assert
	assert.Error(t, err)

	// 驗證所有操作都被回滾，包括第一個成功的操作
	var foundCategory entity.Category
	err = db.First(&foundCategory, "id = ?", "cat-2").Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	var foundProduct entity.Product
	err = db.First(&foundProduct, "id = ?", "prod-2").Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

// TestGetDB 測試 GetDB 輔助函數
func TestGetDB(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}

	t.Run("returns transaction DB when in transaction context", func(t *testing.T) {
		tm := NewGormTransactionManager(db)
		ctx := context.Background()

		err := tm.WithTransaction(ctx, func(txCtx context.Context) error {
			txDB := GetDB(txCtx, db)

			// 驗證返回的是事務 DB
			assert.NotNil(t, txDB)

			// 可以通過檢查 context value 來驗證
			ctxValue := txCtx.Value(repository.TransactionContextKey)
			assert.NotNil(t, ctxValue)
			assert.Equal(t, ctxValue, txDB)

			return nil
		})

		assert.NoError(t, err)
	})

	t.Run("returns default DB when not in transaction context", func(t *testing.T) {
		ctx := context.Background()
		resultDB := GetDB(ctx, db)

		// 驗證返回的是原始 DB with context
		assert.NotNil(t, resultDB)

		// 驗證 context 中沒有事務
		ctxValue := ctx.Value(repository.TransactionContextKey)
		assert.Nil(t, ctxValue)
	})
}
