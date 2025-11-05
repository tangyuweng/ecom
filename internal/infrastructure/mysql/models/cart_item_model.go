package models

import (
	"time"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type CartItemModel struct {
	ID        string       `gorm:"type:char(36);primaryKey"`
	CartID    string       `gorm:"type:char(36);not null;index"`
	ProductID string       `gorm:"type:char(36);not null;index"`
	Quantity  int          `gorm:"type:int;not null;default:1"`
	CreatedAt time.Time    `gorm:"autoCreateTime"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime"`
	Product   ProductModel `gorm:"foreignKey:ProductID;references:ID;constraint:OnDelete:CASCADE"`
}

func (CartItemModel) TableName() string {
	return "cart_items"
}

func (m *CartItemModel) ModelToEntity() *entity.CartItem {
	cartItem := &entity.CartItem{
		ID:        m.ID,
		CartID:    m.CartID,
		ProductID: m.ProductID,
		Quantity:  m.Quantity,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.Product.ID != "" {
		cartItem.Product = m.Product.ModelToEntity()
	}

	return cartItem
}

func (m *CartItemModel) ModelFromEntity(item *entity.CartItem) {
	m.ID = item.ID
	m.CartID = item.CartID
	m.ProductID = item.ProductID
	m.Quantity = item.Quantity
	m.CreatedAt = item.CreatedAt
	m.UpdatedAt = item.UpdatedAt
}
