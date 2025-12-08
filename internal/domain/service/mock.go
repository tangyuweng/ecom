package service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/tangyuweng/ecom/internal/domain/entity"
)

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) NotifyUser(ctx context.Context, userID string, notification *entity.Notification) error {
	args := m.Called(ctx, userID, notification)
	return args.Error(0)
}

func (m *MockNotificationService) Broadcast(ctx context.Context, notification *entity.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationService) GetConnectedUserCount() int {
	args := m.Called()
	return args.Int(0)
}
