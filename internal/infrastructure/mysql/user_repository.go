package mysql

import (
	"context"
	"errors"

	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/repository"
	"github.com/tangyuweng/ecom/internal/infrastructure/mysql/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MysqlUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &MysqlUserRepository{db: db}
}

func (r *MysqlUserRepository) Create(ctx context.Context, user *entity.User) error {
	user.ID = uuid.New().String()

	var model models.UserModel
	model.ModelFromEntity(user)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	user.ID = model.ID
	return nil
}

func (r *MysqlUserRepository) FindAll(ctx context.Context) ([]*entity.User, error) {
	var models []models.UserModel

	err := r.db.WithContext(ctx).Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]*entity.User, len(models))
	for i, model := range models {
		users[i] = model.ModelToEntity()
	}

	return users, nil
}

func (r *MysqlUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var model models.UserModel

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model models.UserModel

	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlUserRepository) FindByAdmin(ctx context.Context) ([]*entity.User, error) {
	var models []models.UserModel

	err := r.db.WithContext(ctx).Where("role = ?", "admin").Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]*entity.User, len(models))
	for i, model := range models {
		users[i] = model.ModelToEntity()
	}

	return users, nil
}

func (r *MysqlUserRepository) Update(ctx context.Context, user *entity.User) error {
	var model models.UserModel
	model.ModelFromEntity(user)
	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *MysqlUserRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.UserModel{}, "id = ?", id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

func (r *MysqlUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		return false, nil
	}

	return count > 0, nil
}
