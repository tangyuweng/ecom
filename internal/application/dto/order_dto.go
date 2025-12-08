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

type OrderListRequest struct {
	Status    *string    `form:"status" binding:"omitempty,oneof=Pending Processing Shipped Completed Cancelled"`
	UserID    *string    `form:"user_id"`
	MinTotal  *float64   `form:"min_total" binding:"omitempty,min=0"`
	MaxTotal  *float64   `form:"max_total" binding:"omitempty,min=0"`
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"end_date" time_format:"2006-01-02"`
	SortBy    string     `form:"sort_by" binding:"omitempty,oneof=order_date total_amount status created_at"`
	SortOrder string     `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Page      int        `form:"page" binding:"omitempty,min=1"`
	PageSize  int        `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type OrderListResponse struct {
	Total  int              `json:"total"`
	Orders []*OrderResponse `json:"orders"`
}
