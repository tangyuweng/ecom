package mysql

import (
	"context"

	"github.com/tangyuweng/ecom/internal/domain/repository"
	"gorm.io/gorm"
)

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) repository.TransactionManager {
	return &GormTransactionManager{db: db}
}

// WithTransaction 在事務中執行函數
// 自動處理 BEGIN、COMMIT、ROLLBACK 和 panic recovery
func (tm *GormTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 處理 panic 和 rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // 重新拋出 panic
		}
	}()

	txCtx := context.WithValue(ctx, repository.TransactionContextKey, tx)

	err := fn(txCtx)
	if err != nil {
		// 業務邏輯返回錯誤，回滾事務
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit().Error
}

// GetDB 從 context 中取得 DB（事務或原始 DB）
// 這是一個輔助函數，供 Repository 使用
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(repository.TransactionContextKey).(*gorm.DB); ok {
		return tx
	}
	return defaultDB.WithContext(ctx)
}
