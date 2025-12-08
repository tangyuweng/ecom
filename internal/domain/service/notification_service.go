package service

import (
	"context"

	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type NotificationService interface {
	NotifyUser(ctx context.Context, userID string, notification *entity.Notification) error
	Broadcast(ctx context.Context, notification *entity.Notification) error
	GetConnectedUserCount() int
}
