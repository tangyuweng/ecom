package websocket

import (
	"context"
	"encoding/json"

	"github.com/tangyuweng/ecom/internal/domain/entity"
	"github.com/tangyuweng/ecom/internal/domain/service"
)

type WsNotificationSvc struct {
	hub *Hub
}

func NewWsNotificationSvc(hub *Hub) service.NotificationService {
	return &WsNotificationSvc{hub: hub}
}

func (s *WsNotificationSvc) NotifyUser(ctx context.Context, userID string, notification *entity.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	s.hub.SendToUser(userID, data)
	return nil
}

func (s *WsNotificationSvc) Broadcast(ctx context.Context, notification *entity.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	s.hub.Broadcast(data)
	return nil
}

func (s *WsNotificationSvc) GetConnectedUserCount() int {
	return s.hub.GetConnectedUserCount()
}
