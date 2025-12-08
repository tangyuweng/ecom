package repository

import "context"

// TransactionManager 管理資料庫事務
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// TransactionKey 用於在 context 中存取事務的 key
type transactionKey struct{}

// TransactionContextKey 是導出的 context key
var TransactionContextKey = transactionKey{}
