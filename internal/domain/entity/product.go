package entity

import (
	"errors"
	"strings"
	"time"
)

type ProductQuery struct {
	Name      *string
	Category  *string
	MinPrice  *float64
	MaxPrice  *float64
	IsActive  *bool
	SortBy    string
	SortOrder string
	Page      int
	PageSize  int
}

type Product struct {
	ID            string
	Name          string
	Description   string
	Price         float64
	StockQuantity int
	IsActive      bool
	CategoryID    string
	CategoryName  string
	CategorySlug  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

var (
	ErrProductNotFound          = errors.New("product not found")
	ErrProductInvalidPrice      = errors.New("product price must be greater than 0")
	ErrProductInvalidStock      = errors.New("product stock quantity cannot be negative")
	ErrProductNameRequired      = errors.New("product name is required")
	ErrProductCategoryRequired  = errors.New("product category is required")
	ErrProductInsufficientStock = errors.New("insufficient stock")
	ErrProductUnauthorized      = errors.New("unauthorized to manage product")
)

func NewProduct(categoryID, name, description string, price float64, stockQuantity int, isActive bool) (*Product, error) {
	if err := validateProductName(name); err != nil {
		return nil, err
	}

	if err := validateProductPrice(price); err != nil {
		return nil, err
	}

	if err := validateProductStockQuantity(stockQuantity); err != nil {
		return nil, err
	}

	now := time.Now()

	return &Product{
		Name:          name,
		Description:   description,
		Price:         price,
		StockQuantity: stockQuantity,
		IsActive:      isActive,
		CategoryID:    categoryID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (p *Product) Update(categoryID, name, description string, price float64, stockQuantity int, isActive bool) error {
	if err := validateProductName(name); err != nil {
		return err
	}

	if err := validateProductPrice(price); err != nil {
		return err
	}

	if err := validateProductStockQuantity(stockQuantity); err != nil {
		return err
	}

	p.Name = name
	p.Description = description
	p.Price = price
	p.StockQuantity = stockQuantity
	p.IsActive = isActive
	p.CategoryID = categoryID
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) UpdateStockQuantity(stockQuantity int) error {
	if err := validateProductStockQuantity(stockQuantity); err != nil {
		return err
	}

	p.StockQuantity = stockQuantity
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) IsStockAvailable() bool {
	return p.StockQuantity > 0
}

func validateProductName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrProductNameRequired
	}
	return nil
}

func validateProductPrice(price float64) error {
	if price < 0 {
		return ErrProductInvalidPrice
	}
	return nil
}

func validateProductStockQuantity(stockQuantity int) error {
	if stockQuantity < 0 {
		return ErrProductInvalidStock
	}
	return nil
}
