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

type MysqlProductRepository struct {
	db *gorm.DB
}

func NewMysqlProductRepository(db *gorm.DB) repository.ProductRepository {
	return &MysqlProductRepository{db: db}
}

func (r *MysqlProductRepository) Create(ctx context.Context, product *entity.Product) error {
	product.ID = uuid.NewString()

	var model models.ProductModel
	model.ModelFromEntity(product)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	product.ID = model.ID
	return nil
}

func (r *MysqlProductRepository) Update(ctx context.Context, product *entity.Product) error {
	var model models.ProductModel
	model.ModelFromEntity(product)
	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *MysqlProductRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.ProductModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrProductNotFound
	}

	return nil
}

func (r *MysqlProductRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	var model models.ProductModel

	err := r.db.WithContext(ctx).Preload("Category").Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrProductNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlProductRepository) FindByCategoryID(ctx context.Context, id string) ([]*entity.Product, error) {
	var models []models.ProductModel

	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("category_id = ?", id).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	products := make([]*entity.Product, len(models))
	for i, model := range models {
		products[i] = model.ModelToEntity()
	}

	return products, nil
}

func (r *MysqlProductRepository) FindByQuery(ctx context.Context, query *entity.ProductQuery) ([]*entity.Product, int, error) {
	var productModels []*models.ProductModel
	var total int64

	db := r.db.WithContext(ctx)

	if query.Name != nil && *query.Name != "" {
		db = db.Where("name LIKE ?", "%"+*query.Name+"%")
	}

	if query.Category != nil && *query.Category != "" {
		db = db.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.slug = ?", *query.Category)
	}

	if query.MinPrice != nil {
		db = db.Where("price >= ?", *query.MinPrice)
	}

	if query.MaxPrice != nil {
		db = db.Where("price <= ?", *query.MaxPrice)
	}

	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	if err := db.Model(&models.ProductModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := query.SortBy + " " + query.SortOrder
	db = db.Order(orderClause)

	offset := (query.Page - 1) * query.PageSize
	db = db.Offset(offset).Limit(query.PageSize)

	if err := db.Preload("Category").Find(&productModels).Error; err != nil {
		return nil, 0, err
	}

	products := make([]*entity.Product, len(productModels))
	for i, model := range productModels {
		products[i] = model.ModelToEntity()
	}

	return products, int(total), nil
}

func (r *MysqlProductRepository) UpdateStock(ctx context.Context, id string, quantity int) error {
	result := r.db.WithContext(ctx).
		Model(&models.ProductModel{}).
		Where("id = ?", id).
		Update("stock_quantity", quantity)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrProductNotFound
	}

	return nil
}

// 刪除 Category 前先檢查是否有商品使用此類別
func (r *MysqlProductRepository) ExistsByCategoryID(ctx context.Context, categoryID string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.ProductModel{}).
		Where("category_id = ?", categoryID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
