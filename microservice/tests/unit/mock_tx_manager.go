package unit

import (
	"context"

	"review-manager/internal/repository"
)

// заглушка TxManager
type MockTxManager struct{}

func (m *MockTxManager) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// compile-time проверка, что MockTxManager реализует интерфейс
var _ repository.TxManager = (*MockTxManager)(nil)
