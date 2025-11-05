package entity

import (
	"errors"
	"time"
)

var (
	ErrCartNotFound       = errors.New("cart not found")
	ErrCartItemNotFound   = errors.New("cart item not found")
	ErrInvalidQuantity    = errors.New("quantity must be greater than 0")
	ErrProductNotInCart   = errors.New("product not in cart")
	ErrEmptyCart          = errors.New("cart is empty")
	ErrCartUnauthorized   = errors.New("unauthorized to access cart")
)

type Cart struct {
	ID        string
	UserID    string
	Items     []*CartItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CartItem struct {
	ID            string
	CartID        string
	ProductID     string
	Quantity      int
	Product       *Product
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewCart(userID string) *Cart {
	now := time.Now()
	return &Cart{
		UserID:    userID,
		Items:     make([]*CartItem, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewCartItem(cartID, productID string, quantity int) (*CartItem, error) {
	if err := validateQuantity(quantity); err != nil {
		return nil, err
	}

	now := time.Now()
	return &CartItem{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (ci *CartItem) UpdateQuantity(quantity int) error {
	if err := validateQuantity(quantity); err != nil {
		return err
	}
	ci.Quantity = quantity
	ci.UpdatedAt = time.Now()
	return nil
}

func (c *Cart) GetTotalItems() int {
	total := 0
	for _, item := range c.Items {
		total += item.Quantity
	}
	return total
}

func (c *Cart) GetTotalPrice() float64 {
	total := 0.0
	for _, item := range c.Items {
		if item.Product != nil {
			total += item.Product.Price * float64(item.Quantity)
		}
	}
	return total
}

func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

func validateQuantity(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	return nil
}
