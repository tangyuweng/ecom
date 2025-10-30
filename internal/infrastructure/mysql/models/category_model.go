package models

import (
	"time"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type CategoryModel struct {
	ID        string    `gorm:"type:char(36);primaryKey"`
	Name      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Slug      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (m *CategoryModel) TableName() string {
	return "categories"
}

func (m *CategoryModel) ModelToEntity() *entity.Category {
	return &entity.Category{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (m *CategoryModel) ModelFromEntity(category *entity.Category) {
	m.ID = category.ID
	m.Name = category.Name
	m.Slug = category.Slug
	m.CreatedAt = category.CreatedAt
	m.UpdatedAt = category.UpdatedAt
}
