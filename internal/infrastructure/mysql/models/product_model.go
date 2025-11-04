package models

import (
	"time"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type ProductModel struct {
	ID            string        `gorm:"type:char(36);primaryKey"`
	Name          string        `gorm:"type:varchar(100);not null;"`
	Description   string        `gorm:"type:text"`
	Price         float64       `gorm:"type:decimal(10,2);not null"`
	StockQuantity int           `gorm:"type:int;not null;default:0"`
	IsActive      bool          `gorm:"type:bool;default:1"`
	CreatedAt     time.Time     `gorm:"autoCreateTime"`
	UpdatedAt     time.Time     `gorm:"autoUpdateTime"`
	CategoryID    string        `gorm:"type:char(36);not null;index"`
	Category      CategoryModel `gorm:"foreignKey:CategoryID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (ProductModel) TableName() string {
	return "products"
}

func (m *ProductModel) ModelToEntity() *entity.Product {
	return &entity.Product{
		ID:            m.ID,
		Name:          m.Name,
		Description:   m.Description,
		Price:         m.Price,
		StockQuantity: m.StockQuantity,
		IsActive:      m.IsActive,
		CategoryID:    m.CategoryID,
		CategoryName:  m.Category.Name,
		CategorySlug:  m.Category.Slug,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func (m *ProductModel) ModelFromEntity(product *entity.Product) {
	m.ID = product.ID
	m.Name = product.Name
	m.Description = product.Description
	m.Price = product.Price
	m.StockQuantity = product.StockQuantity
	m.IsActive = product.IsActive
	m.CategoryID = product.CategoryID
	m.CreatedAt = product.CreatedAt
	m.UpdatedAt = product.UpdatedAt
}
