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

type MysqlCategoryRepository struct {
	db *gorm.DB
}

func NewMysqlCategoryRepository(db *gorm.DB) repository.CategoryRepository {
	return &MysqlCategoryRepository{db: db}
}

func (r *MysqlCategoryRepository) Create(ctx context.Context, category *entity.Category) error {
	db := GetDB(ctx, r.db)
	category.ID = uuid.NewString()

	var model models.CategoryModel
	model.ModelFromEntity(category)

	if err := db.Create(&model).Error; err != nil {
		return err
	}

	category.ID = model.ID
	return nil
}

func (r *MysqlCategoryRepository) FindByID(ctx context.Context, id string) (*entity.Category, error) {
	db := GetDB(ctx, r.db)
	var model models.CategoryModel

	err := db.Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCategoryNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlCategoryRepository) FindByName(ctx context.Context, name string) (*entity.Category, error) {
	db := GetDB(ctx, r.db)
	var model models.CategoryModel

	err := db.Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCategoryNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlCategoryRepository) FindAll(ctx context.Context) ([]*entity.Category, error) {
	db := GetDB(ctx, r.db)
	var models []models.CategoryModel

	err := db.Find(&models).Error
	if err != nil {
		return nil, err
	}

	categories := make([]*entity.Category, len(models))
	for i, model := range models {
		categories[i] = model.ModelToEntity()
	}

	return categories, nil
}

func (r *MysqlCategoryRepository) Update(ctx context.Context, category *entity.Category) error {
	db := GetDB(ctx, r.db)
	var model models.CategoryModel
	model.ModelFromEntity(category)

	return db.Save(&model).Error
}

func (r *MysqlCategoryRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	result := db.Delete(&models.CategoryModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrCategoryNotFound
	}

	return nil
}

func (r *MysqlCategoryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	db := GetDB(ctx, r.db)
	var count int64

	err := db.Model(&models.CategoryModel{}).
		Where("name = ?", name).
		Count(&count).Error

	if err != nil {
		return false, nil
	}

	return count > 0, nil
}
