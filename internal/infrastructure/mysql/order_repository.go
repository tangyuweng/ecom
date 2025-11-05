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

type MysqlOrderRepository struct {
	db *gorm.DB
}

func NewMysqlOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &MysqlOrderRepository{db: db}
}

// Order operations
func (r *MysqlOrderRepository) Create(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order.ID = uuid.NewString()

		var orderModel models.OrderModel
		orderModel.ModelFromEntity(order)

		if err := tx.Create(&orderModel).Error; err != nil {
			return err
		}

		if len(order.Items) > 0 {
			for _, item := range order.Items {
				item.ID = uuid.NewString()
				item.OrderID = order.ID

				var itemModel models.OrderItemModel
				itemModel.ModelFromEntity(item)

				if err := tx.Create(&itemModel).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *MysqlOrderRepository) FindByID(ctx context.Context, orderID string) (*entity.Order, error) {
	var model models.OrderModel

	err := r.db.WithContext(ctx).
		Preload("Items.Product.Category").
		Where("id = ?", orderID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrOrderNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlOrderRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Order, error) {
	var models []models.OrderModel

	err := r.db.WithContext(ctx).
		Preload("Items.Product.Category").
		Where("user_id = ?", userID).
		Order("order_date DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	orders := make([]*entity.Order, len(models))
	for i, model := range models {
		orders[i] = model.ModelToEntity()
	}

	return orders, nil
}

func (r *MysqlOrderRepository) Update(ctx context.Context, order *entity.Order) error {
	var model models.OrderModel
	model.ModelFromEntity(order)

	result := r.db.WithContext(ctx).Model(&model).Updates(map[string]interface{}{
		"status":           order.Status,
		"shipping_address": order.ShippingAddress,
		"recipient_name":   order.RecipientName,
		"updated_at":       order.UpdatedAt,
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrOrderNotFound
	}

	return nil
}

func (r *MysqlOrderRepository) UpdateStatus(ctx context.Context, orderID string, status entity.OrderStatus) error {
	result := r.db.WithContext(ctx).
		Model(&models.OrderModel{}).
		Where("id = ?", orderID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrOrderNotFound
	}

	return nil
}

// OrderItem operations
func (r *MysqlOrderRepository) CreateItems(ctx context.Context, items []*entity.OrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			item.ID = uuid.NewString()

			var model models.OrderItemModel
			model.ModelFromEntity(item)

			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *MysqlOrderRepository) FindItemsByOrderID(ctx context.Context, orderID string) ([]*entity.OrderItem, error) {
	var models []models.OrderItemModel

	err := r.db.WithContext(ctx).
		Preload("Product.Category").
		Where("order_id = ?", orderID).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	items := make([]*entity.OrderItem, len(models))
	for i, model := range models {
		items[i] = model.ModelToEntity()
	}

	return items, nil
}
