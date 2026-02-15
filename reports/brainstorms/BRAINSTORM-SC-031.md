# Brainstorm: SC-031 - Transaction Ingestion API

## 1. Problem Statement
We need to create the entry point for the Fraud Detection System. This API endpoint (`POST /api/v1/transactions`) allows merchants (Tenants) to submit transaction data for real-time risk evaluation.

For this specific task (SC-031), the goal is **Ingestion scaffolding**:
- Create the API route.
- Validate the payload.
- Persist the raw transaction to the database (using `TransactionRepository` created in Epic 2).
- Return a successful response.

**Note**: Actual scoring logic (Rule Engine, ML) comes in later tasks (SC-039+). For now, we prioritize the data pipelines and API contract.

## 2. Requirements

### Functional
1.  **Endpoint**: `POST /api/v1/transactions`
2.  **Auth**: Tenant API Key (extracted via Middleware - mock for now or reuse existing). *Note: SC-029 implementation is in Epic 3 plan but SC-031 depends on it. We might need a simple mock auth or implement basic integration.*
3.  **Input Validation**:
    - `amount`: Required, > 0.
    - `currency`: Required, ISO 3-letter code.
    - `device_id` / `ip_address`: Optional but recommended.
    - `payload`: Arbitrary JSON for merchant-specific fields.
4.  **Persistence**: Save to `transactions` table via `UnitOfWork`.
5.  **Output**: JSON response with `transaction_id`.

### Non-Functional
1.  **Latency**: Minimal overhead (just DB write).
2.  **Error Handling**: Standard HTTP 400 for validation errors, 500 for internal errors.
3.  **Structure**: Follow RESTful conventions.

## 3. API Contract (Draft)

### Request
```json
POST /api/v1/transactions
Content-Type: application/json
X-Tenant-ID: <uuid> (or via Bearer Token)

{
  "user_id": "user_12345",
  "amount": 150.50,
  "currency": "USD",
  "ip_address": "192.168.1.1",
  "device_id": "device_hash_xyz",
  "payload": {
    "item_category": "electronics",
    "shipping_method": "express"
  }
}
```

### Response (Success)
```json
HTTP/1.1 201 Created

{
  "status": "success",
  "data": {
    "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "RECEIVED"
  }
}
```

### Response (Error)
```json
HTTP/1.1 400 Bad Request

{
  "status": "error",
  "message": "Validation failed",
  "errors": [
    {
      "field": "currency",
      "message": "Must be valid ISO 4217 code"
    }
  ]
}
```

## 4. Success Criteria
- [ ] API accepts valid JSON.
- [ ] Validation rejects invalid data (e.g. negative amount).
- [ ] Data is saved to PostgreSQL `transactions` table.
- [ ] Response includes generated `transaction_id`.
