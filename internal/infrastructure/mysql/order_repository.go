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
	db := GetDB(ctx, r.db)

	order.ID = uuid.NewString()

	var orderModel models.OrderModel
	orderModel.ModelFromEntity(order)

	if err := db.Create(&orderModel).Error; err != nil {
		return err
	}

	if len(order.Items) > 0 {
		for _, item := range order.Items {
			item.ID = uuid.NewString()
			item.OrderID = order.ID

			var itemModel models.OrderItemModel
			itemModel.ModelFromEntity(item)

			if err := db.Create(&itemModel).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *MysqlOrderRepository) FindByID(ctx context.Context, orderID string) (*entity.Order, error) {
	db := GetDB(ctx, r.db)
	var model models.OrderModel

	err := db.Preload("Items.Product.Category").
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
	db := GetDB(ctx, r.db)
	var models []models.OrderModel

	err := db.Preload("Items.Product.Category").
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

func (r *MysqlOrderRepository) FindByQuery(ctx context.Context, query *entity.OrderQuery) ([]*entity.Order, int, error) {
	var orderModels []models.OrderModel
	var total int64

	db := GetDB(ctx, r.db)

	// 構建動態查詢條件
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	if query.UserID != nil && *query.UserID != "" {
		db = db.Where("user_id = ?", *query.UserID)
	}

	if query.MinTotal != nil {
		db = db.Where("total_amount >= ?", *query.MinTotal)
	}

	if query.MaxTotal != nil {
		db = db.Where("total_amount <= ?", *query.MaxTotal)
	}

	if query.StartDate != nil {
		db = db.Where("order_date >= ?", *query.StartDate)
	}

	if query.EndDate != nil {
		db = db.Where("order_date <= ?", *query.EndDate)
	}

	// 計算總數
	if err := db.Model(&models.OrderModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 應用排序
	orderClause := query.SortBy + " " + query.SortOrder
	db = db.Order(orderClause)

	// 應用分頁
	offset := (query.Page - 1) * query.PageSize
	db = db.Offset(offset).Limit(query.PageSize)

	// 查詢訂單並預加載關聯數據
	if err := db.Preload("Items.Product.Category").Find(&orderModels).Error; err != nil {
		return nil, 0, err
	}

	orders := make([]*entity.Order, len(orderModels))
	for i, model := range orderModels {
		orders[i] = model.ModelToEntity()
	}

	return orders, int(total), nil
}

func (r *MysqlOrderRepository) Update(ctx context.Context, order *entity.Order) error {
	db := GetDB(ctx, r.db)
	var model models.OrderModel
	model.ModelFromEntity(order)

	result := db.Model(&model).Updates(map[string]interface{}{
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

// OrderItem operations
func (r *MysqlOrderRepository) CreateItems(ctx context.Context, items []*entity.OrderItem) error {
	db := GetDB(ctx, r.db)

	for _, item := range items {
		item.ID = uuid.NewString()

		var model models.OrderItemModel
		model.ModelFromEntity(item)

		if err := db.Create(&model).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *MysqlOrderRepository) FindItemsByOrderID(ctx context.Context, orderID string) ([]*entity.OrderItem, error) {
	db := GetDB(ctx, r.db)
	var models []models.OrderItemModel

	err := db.Preload("Product.Category").
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
