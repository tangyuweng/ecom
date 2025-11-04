package dto

import "time"

type ProductListRequest struct {
	Name      *string  `form:"name"`
	Category  *string  `form:"category"`
	MinPrice  *float64 `form:"min_price"`
	MaxPrice  *float64 `form:"max_price"`
	IsActive  *bool    `form:"is_active"`
	SortBy    string   `form:"sort_by" binding:"omitempty,oneof=price created_at name"`
	SortOrder string   `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Page      int      `form:"page" binding:"omitempty,min=1"`
	PageSize  int      `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type ProductResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	IsActive      bool      `json:"is_active"`
	CategoryID    string    `json:"category_id"`
	CategoryName  string    `json:"category_name"`
	CategorySlug  string    `json:"category_slug"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProductListResponse struct {
	Total    int                `json:"total"`
	Products []*ProductResponse `json:"products"`
}

type CreateProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description" binding:"omitempty"`
	Price         float64 `json:"price" binding:"required,min=0"`
	StockQuantity int     `json:"stock_quantity" binding:"required,min=0"`
	IsActive      bool    `json:"is_active"`
	CategoryID    string  `json:"category_id"`
}

type UpdateProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description" binding:"omitempty"`
	Price         float64 `json:"price" binding:"min=0"`
	StockQuantity int     `json:"stock_quantity" binding:"min=0"`
	IsActive      bool    `json:"is_active"`
	CategoryID    string  `json:"category_id"`
}

type UpdateProductStockRequest struct {
	StockQuantity int `json:"stock_quantity" binding:"required,min=0"`
}
