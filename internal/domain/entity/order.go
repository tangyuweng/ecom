package entity

import (
	"errors"
	"slices"
	"strings"
	"time"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "Pending"
	OrderStatusProcessing OrderStatus = "Processing"
	OrderStatusShipped    OrderStatus = "Shipped"
	OrderStatusCompleted  OrderStatus = "Completed"
	OrderStatusCancelled  OrderStatus = "Cancelled"
)

var (
	ErrOrderNotFound                = errors.New("order not found")
	ErrOrderItemNotFound            = errors.New("order item not found")
	ErrOrderInvalidStatus           = errors.New("invalid order status")
	ErrOrderInvalidStatusTransition = errors.New("invalid order status transition")
	ErrOrderInvalidTotalAmount      = errors.New("total amount must be greater than 0")
	ErrOrderShippingAddressRequired = errors.New("shipping address is required")
	ErrOrderRecipientNameRequired   = errors.New("recipient name is required")
	ErrOrderCannotCancel            = errors.New("order cannot be cancelled")
	ErrOrderUnauthorized            = errors.New("unauthorized to access order")
	ErrOrderEmptyItems              = errors.New("order must contain at least one item")
	ErrOrderInvalidItemPrice        = errors.New("item price must be greater than 0")
	ErrOrderInvalidItemQuantity     = errors.New("item quantity must be greater than 0")
)

type OrderQuery struct {
	Status    *OrderStatus // 訂單狀態過濾
	UserID    *string      // 用戶ID過濾
	MinTotal  *float64     // 最小金額過濾
	MaxTotal  *float64     // 最大金額過濾
	StartDate *time.Time   // 訂單開始日期過濾
	EndDate   *time.Time   // 訂單結束日期過濾
	SortBy    string       // 排序欄位 (order_date, total_amount, status, created_at)
	SortOrder string       // 排序方向 (asc, desc)
	Page      int          // 頁碼
	PageSize  int          // 每頁數量
}

type Order struct {
	ID              string
	UserID          string
	OrderDate       time.Time
	TotalAmount     float64
	Status          OrderStatus
	ShippingAddress string
	RecipientName   string
	Items           []*OrderItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
	UnitPrice float64
	Subtotal  float64
	Product   *Product
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrder(userID, shippingAddress, recipientName string, items []*OrderItem) (*Order, error) {
	if err := validateShippingAddress(shippingAddress); err != nil {
		return nil, err
	}

	if err := validateRecipientName(recipientName); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, ErrOrderEmptyItems
	}

	totalAmount := 0.0
	for _, item := range items {
		if err := validateOrderItem(item); err != nil {
			return nil, err
		}
		totalAmount += item.Subtotal
	}

	now := time.Now()

	return &Order{
		UserID:          userID,
		OrderDate:       now,
		TotalAmount:     totalAmount,
		Status:          OrderStatusPending,
		ShippingAddress: shippingAddress,
		RecipientName:   recipientName,
		Items:           items,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func NewOrderItem(orderID, productID string, quantity int, unitPrice float64) (*OrderItem, error) {
	if err := validateItemQuantity(quantity); err != nil {
		return nil, err
	}

	if err := validateItemPrice(unitPrice); err != nil {
		return nil, err
	}

	subtotal := float64(quantity) * unitPrice
	now := time.Now()

	return &OrderItem{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		UnitPrice: unitPrice,
		Subtotal:  subtotal,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateStatus 更新訂單狀態（僅供管理員使用，需要驗證狀態轉換）
// 允許的狀態轉換：
// Pending -> Processing -> Shipped -> Completed
// Pending -> Cancelled (through Cancel method only)
// Processing -> Cancelled (through Cancel method only)
func (o *Order) UpdateStatus(status OrderStatus) error {
	if !isValidOrderStatus(status) {
		return ErrOrderInvalidStatus
	}

	// 不允許轉換到 Cancelled（必須使用 Cancel 方法）
	if status == OrderStatusCancelled {
		return ErrOrderInvalidStatusTransition
	}

	// 已取消或已完成的訂單不能再變更狀態
	if o.Status == OrderStatusCancelled || o.Status == OrderStatusCompleted {
		return ErrOrderInvalidStatusTransition
	}

	// 驗證狀態轉換是否合法
	if !isValidStatusTransition(o.Status, status) {
		return ErrOrderInvalidStatusTransition
	}

	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

// isValidStatusTransition 檢查狀態轉換是否合法
func isValidStatusTransition(from, to OrderStatus) bool {
	validTransitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:    {OrderStatusProcessing},
		OrderStatusProcessing: {OrderStatusShipped},
		OrderStatusShipped:    {OrderStatusCompleted},
		OrderStatusCompleted:  {}, // 完成後不能轉換
		OrderStatusCancelled:  {}, // 取消後不能轉換
	}

	allowedStatuses, ok := validTransitions[from]
	if !ok {
		return false
	}

	return slices.Contains(allowedStatuses, to)
}

func (o *Order) Cancel() error {
	// Can only cancel orders that are pending or processing
	if o.Status != OrderStatusPending && o.Status != OrderStatusProcessing {
		return ErrOrderCannotCancel
	}

	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusProcessing
}

func (o *Order) IsCompleted() bool {
	return o.Status == OrderStatusCompleted
}

func (o *Order) IsCancelled() bool {
	return o.Status == OrderStatusCancelled
}

func (o *Order) CalculateTotalAmount() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += item.Subtotal
	}
	return total
}

func isValidOrderStatus(status OrderStatus) bool {
	switch status {
	case OrderStatusPending, OrderStatusProcessing, OrderStatusShipped, OrderStatusCompleted, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func validateShippingAddress(address string) error {
	if strings.TrimSpace(address) == "" {
		return ErrOrderShippingAddressRequired
	}
	return nil
}

func validateRecipientName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrOrderRecipientNameRequired
	}
	return nil
}

func validateOrderItem(item *OrderItem) error {
	if err := validateItemQuantity(item.Quantity); err != nil {
		return err
	}
	if err := validateItemPrice(item.UnitPrice); err != nil {
		return err
	}
	return nil
}

func validateItemQuantity(quantity int) error {
	if quantity <= 0 {
		return ErrOrderInvalidItemQuantity
	}
	return nil
}

func validateItemPrice(price float64) error {
	if price <= 0 {
		return ErrOrderInvalidItemPrice
	}
	return nil
}
