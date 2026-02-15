# SCOUT REPORT: SC-032 - Idempotency Implementation

**Date**: 2026-02-15  
**Task**: SC-032  
**Scope**: Add correlation_id support for idempotent transaction creation

---

## 🎯 INTEGRATION POINTS

### 1. Database Schema
**File**: `migrations/000005_add_correlation_id.up.sql` (NEW)

**Current State**:
- `transactions` table exists
- No `correlation_id` column

**Required Changes**:
```sql
ALTER TABLE transactions 
ADD COLUMN correlation_id UUID;

CREATE UNIQUE INDEX idx_transactions_tenant_correlation 
ON transactions (tenant_id, correlation_id) 
WHERE correlation_id IS NOT NULL;
```

---

### 2. Domain Model
**File**: `internal/domain/models/transaction.go`

**Current State**:
```go
type Transaction struct {
    TransactionID uuid.UUID
    TenantID      uuid.UUID
    OccurredAt    time.Time
    Payload       map[string]interface{}
}
```

**Required Changes**:
```go
type Transaction struct {
    TransactionID uuid.UUID
    TenantID      uuid.UUID
    CorrelationID *uuid.UUID  // NEW - nullable
    OccurredAt    time.Time
    Payload       map[string]interface{}
}
```

---

### 3. DTO Layer
**File**: `internal/api/transaction/dto.go`

**Current State**:
```go
type CreateTransactionRequest struct {
    UserID    string
    Amount    float64
    Currency  string
    // ...
}
```

**Required Changes**:
```go
type CreateTransactionRequest struct {
    CorrelationID *uuid.UUID `json:"correlation_id"` // NEW - optional
    UserID        string     `json:"user_id"`
    Amount        float64    `json:"amount"`
    Currency      string     `json:"currency"`
    // ...
}
```

---

### 4. Repository Interface
**File**: `internal/domain/repositories/transaction_repository.go`

**Current State**:
```go
type TransactionRepository interface {
    Save(ctx context.Context, tx *models.Transaction) error
}
```

**Required Changes**:
```go
type TransactionRepository interface {
    Save(ctx context.Context, tx *models.Transaction) error
    FindByCorrelation(ctx context.Context, tenantID, correlationID uuid.UUID) (*models.Transaction, error) // NEW
}
```

---

### 5. Repository Implementation
**File**: `internal/repository/transaction_repository.go`

**Required Changes**:
- Implement `FindByCorrelation` method
- Query: `SELECT * FROM transactions WHERE tenant_id = $1 AND correlation_id = $2`

---

### 6. Service Layer
**File**: `internal/api/transaction/service.go`

**Current Logic**:
```go
func (s *service) Create(ctx, tenantID, req) (*Response, error) {
    // Create transaction directly
    tx := &models.Transaction{...}
    err := s.uow.Transactions().Save(ctx, tx)
    return response, err
}
```

**Required Logic**:
```go
func (s *service) Create(ctx, tenantID, req) (*Response, error) {
    // NEW: Check for existing transaction
    if req.CorrelationID != nil {
        existing, err := s.uow.Transactions().FindByCorrelation(ctx, tenantID, *req.CorrelationID)
        if err == nil {
            // Return existing transaction (idempotent)
            return toResponse(existing), nil
        }
        // If not found, continue to create
    }
    
    // Create new transaction
    tx := &models.Transaction{
        CorrelationID: req.CorrelationID, // NEW
        // ...
    }
    err := s.uow.Transactions().Save(ctx, tx)
    return response, err
}
```

---

### 7. Integration Tests
**File**: `tests/integration/api/transaction_flow_test.go`

**Required Test Cases**:
1. ✅ Create with correlation_id → Success
2. ✅ Duplicate correlation_id → Returns existing transaction
3. ✅ Different tenant, same correlation_id → Creates new
4. ✅ No correlation_id → Works as before (backward compatible)

---

## 📊 FILES TO MODIFY

| File | Type | Changes |
|------|------|---------|
| `migrations/000005_add_correlation_id.up.sql` | NEW | Add column + index |
| `migrations/000005_add_correlation_id.down.sql` | NEW | Rollback migration |
| `internal/domain/models/transaction.go` | MODIFY | Add CorrelationID field |
| `internal/api/transaction/dto.go` | MODIFY | Add CorrelationID to request |
| `internal/domain/repositories/transaction_repository.go` | MODIFY | Add FindByCorrelation method |
| `internal/repository/transaction_repository.go` | MODIFY | Implement FindByCorrelation |
| `internal/api/transaction/service.go` | MODIFY | Add idempotency check |
| `tests/integration/api/transaction_flow_test.go` | MODIFY | Add idempotency tests |

**Total**: 8 files (2 new, 6 modified)

---

## 🔍 DEPENDENCIES

### Required
- ✅ SC-031 (Transaction Ingestion API) - COMPLETE
- ✅ Database connection - AVAILABLE
- ✅ Migration system - AVAILABLE

### Optional
- ⏳ SC-029 (API Key Middleware) - Not required for idempotency

---

## ⚠️ POTENTIAL ISSUES

### 1. Race Conditions
**Issue**: Two concurrent requests with same correlation_id  
**Mitigation**: Unique index will cause one to fail, client retries

### 2. Performance
**Issue**: Extra DB query for every request with correlation_id  
**Mitigation**: Index on (tenant_id, correlation_id) makes query fast

### 3. Backward Compatibility
**Issue**: Existing clients don't send correlation_id  
**Mitigation**: Field is optional (nullable)

---

## ✅ READINESS CHECK

- ✅ Database supports UUID type
- ✅ Repository pattern in place
- ✅ Service layer exists
- ✅ Integration test framework ready
- ✅ Migration system working

**Status**: ✅ **READY TO IMPLEMENT**

---

**Next Step**: Create detailed implementation plan
