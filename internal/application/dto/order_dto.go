package dto

import "time"

type CreateOrderRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required,max=255"`
	RecipientName   string `json:"recipient_name" binding:"required,max=50"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=Pending Processing Shipped Completed Cancelled"`
}

type OrderItemResponse struct {
	ID          string           `json:"id"`
	ProductID   string           `json:"product_id"`
	ProductName string           `json:"product_name"`
	Quantity    int              `json:"quantity"`
	UnitPrice   float64          `json:"unit_price"`
	Subtotal    float64          `json:"subtotal"`
	Product     *ProductResponse `json:"product,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type OrderResponse struct {
	ID              string               `json:"id"`
	UserID          string               `json:"user_id"`
	OrderDate       time.Time            `json:"order_date"`
	TotalAmount     float64              `json:"total_amount"`
	Status          string               `json:"status"`
	ShippingAddress string               `json:"shipping_address"`
	RecipientName   string               `json:"recipient_name"`
	Items           []*OrderItemResponse `json:"items"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}
