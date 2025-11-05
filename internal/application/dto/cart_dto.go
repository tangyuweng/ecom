package dto

import "time"

type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type CartItemResponse struct {
	ID           string           `json:"id"`
	ProductID    string           `json:"product_id"`
	ProductName  string           `json:"product_name"`
	ProductPrice float64          `json:"product_price"`
	Quantity     int              `json:"quantity"`
	Subtotal     float64          `json:"subtotal"`
	Product      *ProductResponse `json:"product,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type CartResponse struct {
	ID         string              `json:"id"`
	UserID     string              `json:"user_id"`
	Items      []*CartItemResponse `json:"items"`
	TotalItems int                 `json:"total_items"`
	TotalPrice float64             `json:"total_price"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}
