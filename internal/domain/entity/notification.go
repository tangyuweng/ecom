package entity

type NotificationType string

const (
	NotificationTypeOrderStatusUpdated NotificationType = "order_status_updated"
	NotificationTypeOrderCancelled     NotificationType = "order_cancelled"
	NotificationTypeNewOrderCreated    NotificationType = "new_order_created"
)

type Notification struct {
	Type    NotificationType       `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

func NewNotification(notificationType NotificationType, payload map[string]interface{}) *Notification {
	return &Notification{
		Type:    notificationType,
		Payload: payload,
	}
}
