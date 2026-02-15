package transaction

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sentinal/core/internal/domain/models"
	"github.com/sentinal/core/internal/domain/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -- Mocks --

type MockUnitOfWork struct {
	mock.Mock
}

func (m *MockUnitOfWork) Do(ctx context.Context, fn func(repositories.UnitOfWork) error) error {
	args := m.Called(ctx, fn)
	// Execute the function with the mock itself to simulate atomic block
	if fn != nil && args.Error(0) == nil {
		return fn(m)
	}
	return args.Error(0)
}

func (m *MockUnitOfWork) Transactions() repositories.TransactionRepository {
	args := m.Called()
	return args.Get(0).(repositories.TransactionRepository)
}

func (m *MockUnitOfWork) Tenants() repositories.TenantRepository {
	return nil
}
func (m *MockUnitOfWork) Users() repositories.UserRepository {
	return nil
}
func (m *MockUnitOfWork) RBAC() repositories.RBACRepository {
	return nil
}

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Save(ctx context.Context, tx *models.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByCorrelation(ctx context.Context, tenantID, correlationID uuid.UUID) (*models.Transaction, error) {
	args := m.Called(ctx, tenantID, correlationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) SaveBulk(ctx context.Context, transactions []*models.Transaction) (int64, error) {
	args := m.Called(ctx, transactions)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

// -- Tests --

func TestCreate_Success_NewTransaction(t *testing.T) {
	mockUoW := new(MockUnitOfWork)
	mockRepo := new(MockTransactionRepository)

	svc := NewService(mockUoW)

	ctx := context.Background()
	tenantID := uuid.New()
	correlationID := uuid.New()

	req := &CreateTransactionRequest{
		CorrelationID: &correlationID,
		UserID:        "user-123",
		Amount:        100.50,
		Currency:      "USD",
		DeviceID:      "device-1",
		IPAddress:     "127.0.0.1",
	}

	// 1. Check for existing (Not found)
	mockUoW.On("Transactions").Return(mockRepo)
	mockRepo.On("FindByCorrelation", ctx, tenantID, correlationID).Return(nil, assert.AnError)

	// 2. Do block / Save
	mockUoW.On("Do", ctx, mock.Anything).Return(nil)
	// Inside Do, Transactions() is called again, effectively covered by the .Return(mockRepo) above if checking strictly times=2
	mockRepo.On("Save", ctx, mock.AnythingOfType("*models.Transaction")).Return(nil)

	resp, err := svc.Create(ctx, tenantID, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "received", resp.Status)
	assert.NotEmpty(t, resp.TransactionID)

	mockRepo.AssertExpectations(t)
	mockUoW.AssertExpectations(t)
}

func TestCreate_Success_Idempotent(t *testing.T) {
	mockUoW := new(MockUnitOfWork)
	mockRepo := new(MockTransactionRepository)

	svc := NewService(mockUoW)

	ctx := context.Background()
	tenantID := uuid.New()
	correlationID := uuid.New()
	existingTxID := uuid.New()

	req := &CreateTransactionRequest{
		CorrelationID: &correlationID,
		UserID:        "user-123",
		Amount:        100.50,
		Currency:      "USD",
	}

	existingTx := &models.Transaction{
		TransactionID: existingTxID,
		TenantID:      tenantID,
		CorrelationID: &correlationID,
		OccurredAt:    time.Now(),
	}

	// 1. Check for existing (Found)
	// We call Transactions() to get repo
	mockUoW.On("Transactions").Return(mockRepo)
	mockRepo.On("FindByCorrelation", ctx, tenantID, correlationID).Return(existingTx, nil)

	// Save should NOT be called
	// Do should NOT be called

	resp, err := svc.Create(ctx, tenantID, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, existingTxID, resp.TransactionID)
	assert.Equal(t, "received", resp.Status)

	mockRepo.AssertExpectations(t)
	mockUoW.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Save")
}
