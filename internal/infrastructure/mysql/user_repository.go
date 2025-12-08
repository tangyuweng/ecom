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
	db := GetDB(ctx, r.db)
	user.ID = uuid.New().String()

	var model models.UserModel
	model.ModelFromEntity(user)

	if err := db.Create(&model).Error; err != nil {
		return err
	}

	user.ID = model.ID
	return nil
}

func (r *MysqlUserRepository) FindAll(ctx context.Context) ([]*entity.User, error) {
	db := GetDB(ctx, r.db)
	var models []models.UserModel

	err := db.Find(&models).Error
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
	db := GetDB(ctx, r.db)
	var model models.UserModel

	err := db.Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	db := GetDB(ctx, r.db)
	var model models.UserModel

	err := db.Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}

	return model.ModelToEntity(), nil
}

func (r *MysqlUserRepository) FindByAdmin(ctx context.Context) ([]*entity.User, error) {
	db := GetDB(ctx, r.db)
	var models []models.UserModel

	err := db.Where("role = ?", "admin").Find(&models).Error
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
	db := GetDB(ctx, r.db)
	var model models.UserModel
	model.ModelFromEntity(user)
	return db.Save(&model).Error
}

func (r *MysqlUserRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	result := db.Delete(&models.UserModel{}, "id = ?", id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

func (r *MysqlUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	db := GetDB(ctx, r.db)
	var count int64

	err := db.Model(&models.UserModel{}).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		return false, nil
	}

	return count > 0, nil
}
