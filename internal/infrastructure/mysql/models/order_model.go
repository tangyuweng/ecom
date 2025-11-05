package models

import (
	"time"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type OrderModel struct {
	ID              string           `gorm:"type:char(36);primaryKey"`
	UserID          string           `gorm:"type:char(36);not null;index"`
	OrderDate       time.Time        `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	TotalAmount     float64          `gorm:"type:decimal(10,2);not null"`
	Status          string           `gorm:"type:enum('Pending','Processing','Shipped','Completed','Cancelled');not null;default:'Pending';index"`
	ShippingAddress string           `gorm:"type:varchar(255);not null"`
	RecipientName   string           `gorm:"type:varchar(50);not null"`
	CreatedAt       time.Time        `gorm:"autoCreateTime"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime"`
	Items           []OrderItemModel `gorm:"foreignKey:OrderID;references:ID;constraint:OnDelete:CASCADE"`
}

func (OrderModel) TableName() string {
	return "orders"
}

func (m *OrderModel) ModelToEntity() *entity.Order {
	order := &entity.Order{
		ID:              m.ID,
		UserID:          m.UserID,
		OrderDate:       m.OrderDate,
		TotalAmount:     m.TotalAmount,
		Status:          entity.OrderStatus(m.Status),
		ShippingAddress: m.ShippingAddress,
		RecipientName:   m.RecipientName,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		Items:           make([]*entity.OrderItem, len(m.Items)),
	}

	for i, item := range m.Items {
		order.Items[i] = item.ModelToEntity()
	}

	return order
}

func (m *OrderModel) ModelFromEntity(order *entity.Order) {
	m.ID = order.ID
	m.UserID = order.UserID
	m.OrderDate = order.OrderDate
	m.TotalAmount = order.TotalAmount
	m.Status = string(order.Status)
	m.ShippingAddress = order.ShippingAddress
	m.RecipientName = order.RecipientName
	m.CreatedAt = order.CreatedAt
	m.UpdatedAt = order.UpdatedAt
}
