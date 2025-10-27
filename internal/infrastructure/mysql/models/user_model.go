package models

import (
	"time"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type UserModel struct {
	ID        string    `gorm:"type:char(36);primaryKey"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
	Name      string    `gorm:"type:varchar(50);not null"`
	Phone     string    `gorm:"type:varchar(20)"`
	Role      string    `gorm:"type:varchar(20);not null;default:'user'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (UserModel) TableName() string {
	return "users"
}

func (m *UserModel) ModelToEntity() *entity.User {
	return &entity.User{
		ID:        m.ID,
		Email:     m.Email,
		Password:  m.Password,
		Name:      m.Name,
		Phone:     m.Phone,
		Role:      entity.UserRole(m.Role),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (m *UserModel) ModelFromEntity(user *entity.User) {
	m.ID = user.ID
	m.Email = user.Email
	m.Password = user.Password
	m.Name = user.Name
	m.Phone = user.Phone
	m.Role = string(user.Role)
	m.CreatedAt = user.CreatedAt
	m.UpdatedAt = user.UpdatedAt
}
