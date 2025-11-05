package models

import (
	"time"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type OrderItemModel struct {
	ID        string       `gorm:"type:char(36);primaryKey"`
	OrderID   string       `gorm:"type:char(36);not null;index"`
	ProductID string       `gorm:"type:char(36);not null;index"`
	Quantity  int          `gorm:"type:int;not null"`
	UnitPrice float64      `gorm:"type:decimal(10,2);not null"`
	Subtotal  float64      `gorm:"type:decimal(10,2);->;generated"`
	CreatedAt time.Time    `gorm:"autoCreateTime"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime"`
	Product   ProductModel `gorm:"foreignKey:ProductID;references:ID;constraint:OnDelete:CASCADE"`
}

func (OrderItemModel) TableName() string {
	return "order_items"
}

func (m *OrderItemModel) ModelToEntity() *entity.OrderItem {
	orderItem := &entity.OrderItem{
		ID:        m.ID,
		OrderID:   m.OrderID,
		ProductID: m.ProductID,
		Quantity:  m.Quantity,
		UnitPrice: m.UnitPrice,
		Subtotal:  m.Subtotal,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.Product.ID != "" {
		orderItem.Product = m.Product.ModelToEntity()
	}

	return orderItem
}

func (m *OrderItemModel) ModelFromEntity(item *entity.OrderItem) {
	m.ID = item.ID
	m.OrderID = item.OrderID
	m.ProductID = item.ProductID
	m.Quantity = item.Quantity
	m.UnitPrice = item.UnitPrice
	m.Subtotal = item.Subtotal
	m.CreatedAt = item.CreatedAt
	m.UpdatedAt = item.UpdatedAt
}
