package models

import (
	"time"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type CartModel struct {
	ID        string          `gorm:"type:char(36);primaryKey"`
	UserID    string          `gorm:"type:char(36);not null;uniqueIndex"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime"`
	Items     []CartItemModel `gorm:"foreignKey:CartID;references:ID;constraint:OnDelete:CASCADE"`
}

func (CartModel) TableName() string {
	return "carts"
}

func (m *CartModel) ModelToEntity() *entity.Cart {
	cart := &entity.Cart{
		ID:        m.ID,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		Items:     make([]*entity.CartItem, len(m.Items)),
	}

	for i, item := range m.Items {
		cart.Items[i] = item.ModelToEntity()
	}

	return cart
}

func (m *CartModel) ModelFromEntity(cart *entity.Cart) {
	m.ID = cart.ID
	m.UserID = cart.UserID
	m.CreatedAt = cart.CreatedAt
	m.UpdatedAt = cart.UpdatedAt
}
