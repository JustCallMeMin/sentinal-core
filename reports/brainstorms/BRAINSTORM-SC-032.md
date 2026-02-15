# BRAINSTORM: SC-032 - Implement Idempotency

**Date**: 2026-02-15  
**Task**: SC-032  
**Goal**: Add idempotency support via `correlation_id`

---

## 🎯 PROBLEM STATEMENT

**Current State**:
- Transaction API accepts duplicate requests
- No way to prevent duplicate transaction creation
- Client retries create multiple transactions

**Desired State**:
- Clients can safely retry requests
- Duplicate `correlation_id` returns existing transaction
- No duplicate transactions in database

---

## 💡 SOLUTION APPROACH

### Option 1: correlation_id in Request Body (Recommended)
```json
POST /api/v1/transactions
{
  "correlation_id": "client-generated-uuid",
  "user_id": "user_123",
  "amount": 100.00,
  "currency": "USD"
}
```

**Pros**:
- ✅ Client controls uniqueness
- ✅ Simple to implement
- ✅ Standard pattern (Stripe, PayPal use this)

**Cons**:
- ⚠️ Requires client to generate UUID

### Option 2: Idempotency-Key Header
```
POST /api/v1/transactions
Idempotency-Key: client-generated-uuid
```

**Pros**:
- ✅ Separates business data from idempotency
- ✅ RESTful pattern

**Cons**:
- ⚠️ More complex (need header parsing)
- ⚠️ Less visible in request body

**Decision**: Use **Option 1** (correlation_id in body) for simplicity.

---

## 🏗️ IMPLEMENTATION PLAN

### 1. Database Schema
```sql
-- Add correlation_id to transactions table
ALTER TABLE transactions 
ADD COLUMN correlation_id UUID;

-- Add unique index for idempotency
CREATE UNIQUE INDEX idx_transactions_tenant_correlation 
ON transactions (tenant_id, correlation_id) 
WHERE correlation_id IS NOT NULL;
```

### 2. DTO Changes
```go
type CreateTransactionRequest struct {
    CorrelationID *uuid.UUID      `json:"correlation_id"` // Optional
    UserID        string          `json:"user_id"`
    Amount        float64         `json:"amount"`
    Currency      string          `json:"currency"`
    // ...
}
```

### 3. Service Logic
```go
func (s *service) Create(ctx, tenantID, req) (*Response, error) {
    // If correlation_id provided, check for existing
    if req.CorrelationID != nil {
        existing, err := s.uow.Transactions().FindByCorrelation(ctx, tenantID, *req.CorrelationID)
        if err == nil {
            // Found existing - return it (idempotent)
            return &CreateTransactionResponse{
                TransactionID: existing.TransactionID,
                Status:        "received",
                ReceivedAt:    existing.OccurredAt,
            }, nil
        }
    }
    
    // Create new transaction
    // ...
}
```

### 4. Repository Method
```go
// Add to TransactionRepository interface
FindByCorrelation(ctx context.Context, tenantID, correlationID uuid.UUID) (*models.Transaction, error)
```

---

## 🧪 TEST SCENARIOS

### Success Cases
1. ✅ **First Request**: correlation_id=A → Creates transaction T1
2. ✅ **Duplicate Request**: correlation_id=A → Returns existing T1 (same response)
3. ✅ **No correlation_id**: → Creates new transaction (backward compatible)

### Edge Cases
1. ⚠️ **Different tenant, same correlation_id**: → Creates new transaction (tenant isolation)
2. ⚠️ **Null correlation_id**: → Treated as no correlation_id

---

## 📊 API CONTRACT

### Request
```json
{
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000", // Optional
  "user_id": "user_123",
  "amount": 100.00,
  "currency": "USD"
}
```

### Response (First Request)
```json
{
  "transaction_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "status": "received",
  "received_at": "2026-02-15T10:00:00Z"
}
```

### Response (Duplicate Request)
```json
{
  "transaction_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7", // Same ID
  "status": "received",
  "received_at": "2026-02-15T10:00:00Z" // Original timestamp
}
```

**Status Code**: 201 (both first and duplicate return 201 for consistency)

---

## 🚨 RISKS & MITIGATIONS

| Risk | Mitigation |
|------|------------|
| Race condition (concurrent requests) | Unique index prevents duplicates |
| Performance (extra DB query) | Index on (tenant_id, correlation_id) |
| Breaking change | Make correlation_id optional |

---

## ✅ ACCEPTANCE CRITERIA

1. ✅ correlation_id field added to schema
2. ✅ Unique index prevents duplicates
3. ✅ Service checks for existing transaction
4. ✅ Returns existing transaction if found
5. ✅ Integration test for idempotency
6. ✅ Backward compatible (correlation_id optional)

---

**Decision**: Proceed with Option 1 (correlation_id in request body)  
**Next**: Scout existing code for integration points
