package service

import "context"

type NotificationService interface {
	NotifyUser(ctx context.Context, userID string, notification *Notification) error
	Broadcast(ctx context.Context, notification *Notification) error
	GetConnectedUserCount() int
}

type Notification struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

func NewNotification(notificationType string, payload map[string]interface{}) *Notification {
	return &Notification{
		Type:    notificationType,
		Payload: payload,
	}
}

const (
	NotificationTypeOrderStatusUpdated = "order_status_updated"
	NotificationTypeOrderCancelled     = "order_cancelled"
)
