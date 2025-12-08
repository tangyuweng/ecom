package mysql

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
	"github.com/tangyuweng/ecom/internal/infrastructure/mysql/models"
	"gorm.io/gorm"
)

type MysqlCartRepository struct {
	db *gorm.DB
}

func NewMysqlCartRepository(db *gorm.DB) repository.CartRepository {
	return &MysqlCartRepository{db: db}
}

// Cart operations
func (r *MysqlCartRepository) FindByUserID(ctx context.Context, userID string) (*entity.Cart, error) {
	db := GetDB(ctx, r.db)
	var model models.CartModel

	err := db.Preload("Items.Product.Category").
		Where("user_id = ?", userID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCartNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlCartRepository) Create(ctx context.Context, cart *entity.Cart) error {
	db := GetDB(ctx, r.db)
	cart.ID = uuid.NewString()

	var model models.CartModel
	model.ModelFromEntity(cart)

	if err := db.Create(&model).Error; err != nil {
		return err
	}

	cart.ID = model.ID
	return nil
}

func (r *MysqlCartRepository) Delete(ctx context.Context, userID string) error {
	db := GetDB(ctx, r.db)

	var cart models.CartModel
	err := db.Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.ErrCartNotFound
		}
		return err
	}

	// Delete all cart items first (CASCADE should handle this, but being explicit)
	// if err := db.Where("cart_id = ?", cart.ID).Delete(&models.CartItemModel{}).Error; err != nil {
	// 	return err
	// }

	return db.Delete(&cart).Error
}

// CartItem operations
func (r *MysqlCartRepository) AddItem(ctx context.Context, item *entity.CartItem) error {
	db := GetDB(ctx, r.db)
	item.ID = uuid.NewString()

	var model models.CartItemModel
	model.ModelFromEntity(item)

	if err := db.Create(&model).Error; err != nil {
		return err
	}

	item.ID = model.ID
	return nil
}

func (r *MysqlCartRepository) UpdateItem(ctx context.Context, item *entity.CartItem) error {
	db := GetDB(ctx, r.db)
	var model models.CartItemModel
	model.ModelFromEntity(item)

	result := db.Model(&model).Updates(map[string]interface{}{
		"quantity":   item.Quantity,
		"updated_at": item.UpdatedAt,
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrCartItemNotFound
	}

	return nil
}

func (r *MysqlCartRepository) RemoveItem(ctx context.Context, cartItemID string) error {
	db := GetDB(ctx, r.db)
	result := db.Delete(&models.CartItemModel{}, "id = ?", cartItemID)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrCartItemNotFound
	}

	return nil
}

func (r *MysqlCartRepository) FindItemByID(ctx context.Context, cartItemID string) (*entity.CartItem, error) {
	db := GetDB(ctx, r.db)
	var model models.CartItemModel

	err := db.Preload("Product.Category").
		Where("id = ?", cartItemID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCartItemNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlCartRepository) FindItemByCartAndProduct(ctx context.Context, cartID, productID string) (*entity.CartItem, error) {
	db := GetDB(ctx, r.db)
	var model models.CartItemModel

	err := db.Preload("Product.Category").
		Where("cart_id = ? AND product_id = ?", cartID, productID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCartItemNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}
