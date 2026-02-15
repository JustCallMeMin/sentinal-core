# PLAN: SC-031 - Implement Transaction Ingestion API

> **Goal**: Scaffold the `POST /api/v1/transactions` endpoint to ingest raw transaction data.

## 1. Domain Object Definitions
- [ ] **Define DTOs**: Create `internal/api/transaction/dto.go`
  - `CreateTransactionRequest`: JSON structure with validation tags (`validate:"required,gt=0"`).
  - `CreateTransactionResponse`: Simple success payload.

## 2. Server Refactoring
- [ ] **Extract Routes**: Create `internal/server/routes.go`
  - Move health/readiness checks from `server.go`.
  - Add `RegisterRoutes()` method to `Server` struct.
  - Call `RegisterRoutes()` in `New()`.

## 3. Service Layer Implementation
- [ ] **Create Service Interface**: `internal/api/transaction/service.go`
  - Interface: `Service`. Method: `Create(ctx, req) (ID, error)`.
  - Implementation: `service`. Logic: Map DTO to Model -> Call Repository via UoW.
- [ ] **Unit Test Service**: `internal/api/transaction/service_test.go` (Mock Repo).

## 4. Handler Layer Implementation
- [ ] **Create Handler**: `internal/api/transaction/handler.go`
  - Struct: `Handler` with dependency on `Service`.
  - Method: `HandleCreate(c *fiber.Ctx) error`.
  - Logic: Parse Body -> Validate -> Extract Tenant ID -> Call Service -> Respond 201.

## 5. Wiring & Integration
- [ ] **Update Server Struct**: Add `TransactionService` to `Server` struct in `internal/server/server.go`.
- [ ] **Register Route**: Add `v1.Post("/transactions", ...)` in `routes.go`.
- [ ] **Main Wiring**: Initialize Service & Handler in `cmd/api/main.go` and inject into Server.

## 6. Verification
- [ ] **Integration Test**: `tests/integration/api/transaction_flow_test.go`
  - Spin up Testcontainers.
  - Execute HTTP Request.
  - Verify DB persistence.
