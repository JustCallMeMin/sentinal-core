# IMPLEMENTATION PLAN: SC-032 - Idempotency

**Date**: 2026-02-15  
**Task**: SC-032  
**Estimated Time**: 0.5 days (~4 hours)

---

## 🎯 OBJECTIVE

Add idempotency support to Transaction Ingestion API via `correlation_id` field.

**Success Criteria**:
1. ✅ Clients can provide optional `correlation_id`
2. ✅ Duplicate `correlation_id` returns existing transaction
3. ✅ Unique index prevents duplicate transactions
4. ✅ Backward compatible (correlation_id optional)
5. ✅ Integration tests verify idempotency

---

## 📋 EXECUTION STEPS

### Step 1: Database Migration (30 min)

#### 1.1 Create Migration Files
```bash
# Create migration
migrate create -ext sql -dir migrations -seq add_correlation_id
```

#### 1.2 Write Up Migration
**File**: `migrations/000005_add_correlation_id.up.sql`
```sql
-- Add correlation_id column (nullable for backward compatibility)
ALTER TABLE transactions 
ADD COLUMN correlation_id UUID;

-- Create unique index for idempotency
-- Partial index (WHERE correlation_id IS NOT NULL) for efficiency
CREATE UNIQUE INDEX idx_transactions_tenant_correlation 
ON transactions (tenant_id, correlation_id) 
WHERE correlation_id IS NOT NULL;

-- Add comment
COMMENT ON COLUMN transactions.correlation_id IS 'Client-provided idempotency key';
```

#### 1.3 Write Down Migration
**File**: `migrations/000005_add_correlation_id.down.sql`
```sql
DROP INDEX IF EXISTS idx_transactions_tenant_correlation;
ALTER TABLE transactions DROP COLUMN IF EXISTS correlation_id;
```

#### 1.4 Apply Migration
```bash
make migrate-up
```

**Verification**:
```sql
\d transactions  -- Should show correlation_id column
\di              -- Should show idx_transactions_tenant_correlation
```

---

### Step 2: Update Domain Model (15 min)

#### 2.1 Update Transaction Model
**File**: `internal/domain/models/transaction.go`

```go
type Transaction struct {
    TransactionID uuid.UUID
    TenantID      uuid.UUID
    CorrelationID *uuid.UUID              // NEW - nullable
    OccurredAt    time.Time
    Payload       map[string]interface{}
}
```

**Note**: Use pointer `*uuid.UUID` for nullable field.

---

### Step 3: Update DTO (15 min)

#### 3.1 Update Request DTO
**File**: `internal/api/transaction/dto.go`

```go
type CreateTransactionRequest struct {
    CorrelationID *uuid.UUID      `json:"correlation_id"` // NEW - optional
    UserID        string          `json:"user_id"`
    Amount        float64         `json:"amount"`
    Currency      string          `json:"currency"`
    DeviceID      string          `json:"device_id"`
    IPAddress     string          `json:"ip_address"`
    Payload       json.RawMessage `json:"payload"`
}

// Validation - correlation_id is optional, no validation needed
```

**No changes to Response DTO** (already has transaction_id, status, received_at).

---

### Step 4: Update Repository (45 min)

#### 4.1 Update Interface
**File**: `internal/domain/repositories/transaction_repository.go`

```go
type TransactionRepository interface {
    Save(ctx context.Context, tx *models.Transaction) error
    FindByCorrelation(ctx context.Context, tenantID, correlationID uuid.UUID) (*models.Transaction, error) // NEW
}
```

#### 4.2 Implement FindByCorrelation
**File**: `internal/repository/transaction_repository.go`

```go
func (r *transactionRepository) FindByCorrelation(
    ctx context.Context,
    tenantID, correlationID uuid.UUID,
) (*models.Transaction, error) {
    query := `
        SELECT transaction_id, tenant_id, correlation_id, occurred_at, payload
        FROM transactions
        WHERE tenant_id = $1 AND correlation_id = $2
        LIMIT 1
    `
    
    var tx models.Transaction
    var payloadJSON []byte
    
    err := r.pool.QueryRow(ctx, query, tenantID, correlationID).Scan(
        &tx.TransactionID,
        &tx.TenantID,
        &tx.CorrelationID,
        &tx.OccurredAt,
        &payloadJSON,
    )
    
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("transaction not found")
        }
        return nil, fmt.Errorf("query failed: %w", err)
    }
    
    // Unmarshal payload
    if err := json.Unmarshal(payloadJSON, &tx.Payload); err != nil {
        return nil, fmt.Errorf("payload unmarshal failed: %w", err)
    }
    
    return &tx, nil
}
```

#### 4.3 Update Save Method
**File**: `internal/repository/transaction_repository.go`

Update INSERT query to include `correlation_id`:
```go
query := `
    INSERT INTO transactions (transaction_id, tenant_id, correlation_id, occurred_at, payload)
    VALUES ($1, $2, $3, $4, $5)
`
```

---

### Step 5: Update Service Layer (30 min)

#### 5.1 Add Idempotency Check
**File**: `internal/api/transaction/service.go`

```go
func (s *service) Create(
    ctx context.Context,
    tenantID uuid.UUID,
    req *CreateTransactionRequest,
) (*CreateTransactionResponse, error) {
    // NEW: Idempotency check
    if req.CorrelationID != nil {
        existing, err := s.uow.Transactions().FindByCorrelation(ctx, tenantID, *req.CorrelationID)
        if err == nil {
            // Found existing transaction - return it (idempotent)
            return &CreateTransactionResponse{
                TransactionID: existing.TransactionID,
                Status:        "received",
                ReceivedAt:    existing.OccurredAt,
            }, nil
        }
        // If not found (err != nil), continue to create new transaction
    }
    
    // Create new transaction
    tx := &models.Transaction{
        TransactionID: uuid.New(),
        TenantID:      tenantID,
        CorrelationID: req.CorrelationID, // NEW - may be nil
        OccurredAt:    time.Now().UTC(),
        Payload:       buildPayload(req),
    }
    
    err := s.uow.Do(ctx, func(u repositories.UnitOfWork) error {
        return u.Transactions().Save(ctx, tx)
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to save transaction: %w", err)
    }
    
    return &CreateTransactionResponse{
        TransactionID: tx.TransactionID,
        Status:        "received",
        ReceivedAt:    tx.OccurredAt,
    }, nil
}
```

---

### Step 6: Integration Tests (45 min)

#### 6.1 Add Test Cases
**File**: `tests/integration/api/transaction_flow_test.go`

```go
t.Run("Idempotency - First Request", func(t *testing.T) {
    correlationID := uuid.New()
    payload := map[string]interface{}{
        "correlation_id": correlationID.String(),
        "user_id":        "user_123",
        "amount":         100.00,
        "currency":       "USD",
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Tenant-ID", testTenant.TenantID.String())
    
    resp, err := srv.App.Test(req)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, 201, resp.StatusCode)
    
    var result transaction.CreateTransactionResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    require.NoError(t, err)
    
    // Save for next test
    firstTransactionID := result.TransactionID
})

t.Run("Idempotency - Duplicate Request", func(t *testing.T) {
    // Same correlation_id as previous test
    correlationID := uuid.New() // Use same ID from previous test
    payload := map[string]interface{}{
        "correlation_id": correlationID.String(),
        "user_id":        "user_123",
        "amount":         100.00,
        "currency":       "USD",
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Tenant-ID", testTenant.TenantID.String())
    
    resp, err := srv.App.Test(req)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, 201, resp.StatusCode)
    
    var result transaction.CreateTransactionResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    require.NoError(t, err)
    
    // Should return same transaction ID
    assert.Equal(t, firstTransactionID, result.TransactionID)
})

t.Run("Idempotency - Different Tenant Same Correlation", func(t *testing.T) {
    // Create second tenant
    tenant2 := &models.Tenant{Name: "Tenant 2", IndustrySegment: "Retail"}
    err := uow.Do(ctx, func(u repositories.UnitOfWork) error {
        return u.Tenants().Create(ctx, tenant2)
    })
    require.NoError(t, err)
    
    // Use same correlation_id but different tenant
    correlationID := uuid.New()
    payload := map[string]interface{}{
        "correlation_id": correlationID.String(),
        "user_id":        "user_456",
        "amount":         200.00,
        "currency":       "USD",
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Tenant-ID", tenant2.TenantID.String())
    
    resp, err := srv.App.Test(req)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, 201, resp.StatusCode)
    
    // Should create NEW transaction (different tenant)
    var result transaction.CreateTransactionResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    require.NoError(t, err)
    assert.NotEqual(t, firstTransactionID, result.TransactionID)
})

t.Run("Backward Compatibility - No Correlation ID", func(t *testing.T) {
    payload := map[string]interface{}{
        "user_id":  "user_789",
        "amount":   300.00,
        "currency": "USD",
    }
    
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Tenant-ID", testTenant.TenantID.String())
    
    resp, err := srv.App.Test(req)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, 201, resp.StatusCode)
})
```

---

### Step 7: Update Postman Collection (15 min)

#### 7.1 Add Idempotency Test
**File**: `postman/sentinal-core-smoke-test.postman_collection.json`

Add new request:
```json
{
  "name": "Create Transaction - With Correlation ID",
  "request": {
    "method": "POST",
    "header": [
      {"key": "Content-Type", "value": "application/json"},
      {"key": "X-Tenant-ID", "value": "{{tenant_id}}"}
    ],
    "body": {
      "mode": "raw",
      "raw": "{\n  \"correlation_id\": \"{{$guid}}\",\n  \"user_id\": \"user_123\",\n  \"amount\": 100.00,\n  \"currency\": \"USD\"\n}"
    },
    "url": {
      "raw": "{{base_url}}/api/v1/transactions",
      "host": ["{{base_url}}"],
      "path": ["api", "v1", "transactions"]
    }
  }
}
```

---

## ✅ VERIFICATION CHECKLIST

### Build & Test
- [ ] `go build` succeeds
- [ ] `go test ./...` all pass
- [ ] Integration tests pass (4 new tests)
- [ ] Postman collection updated

### Database
- [ ] Migration applied successfully
- [ ] Index created
- [ ] Can query by correlation_id

### Functionality
- [ ] First request creates transaction
- [ ] Duplicate correlation_id returns existing
- [ ] Different tenant isolation works
- [ ] No correlation_id works (backward compatible)

---

## 📊 ESTIMATED TIME BREAKDOWN

| Step | Task | Time |
|------|------|------|
| 1 | Database Migration | 30 min |
| 2 | Domain Model | 15 min |
| 3 | DTO Update | 15 min |
| 4 | Repository | 45 min |
| 5 | Service Layer | 30 min |
| 6 | Integration Tests | 45 min |
| 7 | Postman Collection | 15 min |
| **Total** | | **~3 hours** |

**Buffer**: +1 hour for debugging/testing  
**Total Estimate**: **4 hours (0.5 days)**

---

## 🚀 EXECUTION ORDER

1. ✅ Create migration files
2. ✅ Apply migration
3. ✅ Update domain model
4. ✅ Update DTO
5. ✅ Update repository interface
6. ✅ Implement FindByCorrelation
7. ✅ Update Save method
8. ✅ Update service logic
9. ✅ Write integration tests
10. ✅ Run tests
11. ✅ Update Postman collection
12. ✅ Final verification

---

**Ready to Execute**: ✅ YES  
**Next**: Begin implementation (Step 1)
