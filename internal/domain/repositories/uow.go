package repositories

import (
	"context"
)

// UnitOfWork interface defines the contract for atomic database operations
type UnitOfWork interface {
	Do(ctx context.Context, fn func(uow UnitOfWork) error) error
	Tenants() TenantRepository
	Transactions() TransactionRepository
	Users() UserRepository
	RBAC() RBACRepository
}
