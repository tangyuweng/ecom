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

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateAccessToken(userID string, version int) (string, error) {
	args := m.Called(userID, version)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GenerateRefreshToken(userID string, version int) (string, error) {
	args := m.Called(userID, version)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(ctx context.Context, token string) (*TokenDetails, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenDetails), args.Error(1)
}

func (m *MockJWTService) ValidateTokenBasic(token string) (*TokenDetails, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenDetails), args.Error(1)
}
