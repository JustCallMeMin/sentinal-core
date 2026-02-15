# Scout Report: SC-031 - Transaction Ingestion API

## 1. System Context
- **Framework**: Go Fiber v2.
- **Middleware**: `TraceMiddleware` (logging), `TransactionMiddleware` (UoW).
- **Existing Routes**: Health checks defined inline in `server.go`.

## 2. Implementation Strategy

### A. Structure Refactoring
We need to move away from inline route definition.
- **New File**: `internal/server/routes.go`
  - Function: `func (s *Server) RegisterRoutes()`
  - Purpose: Group all route definitions (API v1, Health, etc.).

### B. Business Logic Placement
We should follow a Clean Architecture / Domain-Driven Design approach within the constraints of the project.
- **Location**: `internal/api/transaction/` (New directory)
- **Components**:
  - `dto.go`: Structs for Input/Output (Validation tags).
  - `handler.go`: Fiber Specific logic (Parse body, Validate, Call Service, Format Response).
  - `service.go`: Business Logic (Enrichment, Repository Interaction). *Note: For SC-031, this is simple pass-through to repo.*

### C. Dependency Injection
- Update `Server` struct to include `TransactionHandler` (or Service).
- Initialize these in `cmd/api/main.go` and pass to `server.New()`.

## 3. Integration Points
- **Repo**: `internal/repository/uow.go` -> `UnitOfWork` interface.
- **Models**: `internal/domain/models/transaction.go` (Existing model).
- **Middleware**: `TransactionMiddleware` will handle the DB transaction scope automatically for `POST` requests.

## 4. Risks & Constraints
- **Validation**: Ensure strict validation on `amount` and `currency` to prevent bad data.
- **Tenant Context**: Since we don't have full Auth middleware (SC-029) yet, we might need to mock or extract `X-Tenant-ID` header manually in the handler for now.

## 5. Recommendation
- Extract `routes.go` first.
- Create `internal/api/transaction` package.
- Wire everything in `main.go`.
