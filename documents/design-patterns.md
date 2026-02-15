# Sentinal Core - Design Patterns & Anti-Patterns

This document defines the architectural patterns that ensure our code remains maintainable and scalable.

---

## 1. Data Access: Repository Pattern
We use **Interface-Based Repositories** to decouple Business Logic from Infrastructure.

```go
// internal/domain/transaction/repository.go
type Repository interface {
    Get(ctx context.Context, id uuid.UUID) (*Transaction, error)
    List(ctx context.Context, opts ListOptions) ([]*Transaction, error)
}
```

**Anti-Pattern**: Using `pgxpool.Pool` directly in Service Layer.

---

## 2. Business Logic: Strategy Pattern (Rule Engine)
We use the **Comparator Strategy** to evaluate Rules.

```go
// Example
type Operator interface {
    Evaluate(left, right any) (bool, error)
}

// Operators: GreaterThan, LessThan, EqualTo
ruleEngine.Register(">", &GreaterThan{})
```

**Anti-Pattern**: Giant `switch/case` statement for rule logic.

---

## 3. Configuration: Functional Options
We use **Functional Options** for complex constructors.

```go
// Example
func NewServer(opts ...Option) *Server { ... }

db := NewServer(
    WithPort(8080),
    WithDB(conn),
)
```

**Anti-Pattern**: Config struct with default values mixed in logic.

---

## 4. Distributed Systems: Saga Pattern (Future)
When Transactions involve multiple services (Scoring -> Notification -> Audit).

- Current implementation uses **Outbox Pattern** (Transactional Messaging) via Kafka/Redis.
- **Failures must be compensated** (e.g., if Notification fails, mark review task as 'retry').

---

## 5. Domain Modeling: Factory Method
Use Constructors to enforce invariants.

```go
// Example
func NewTransaction(amount float64, currency string) (*Transaction, error) {
    if amount <= 0 { return nil, ErrInvalidAmount }
    return &Transaction{ ... }, nil
}
```

**Anti-Pattern**: Creating structs directly in literal form (`&Transaction{Amount: -100}`) bypassing validation.

---

## 6. Adapter Pattern (Clean Architecture)
Ports (Interfaces) and Adapters (Implementations).

- `cmd/api` -> depends on -> `internal/server` (Port)
- `internal/server` -> depends on -> `internal/platform/postgres` (Adapter)

This allows us to swap Postgres with SQLite for tests effortlessly.
