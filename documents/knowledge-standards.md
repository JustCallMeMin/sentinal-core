# Sentinal Core - Coding Standards & Guidelines

> **Philosophy**: Simple, Explicit, and Boring.
> "Clear is better than clever." - Rob Pike

---

## 1. Commit Convention (Strict)
We follow **Conventional Commits v1.0.0**.

**Structure**:
```
type(scope): subject

body

footer
```

**Types**:
- `feat`: New feature (CORRELATES with MINOR version).
- `fix`: Bug fix (CORRELATES with PATCH version).
- `docs`: Documentation only.
- `style`: Formatting, missing semi colons, etc; no code change.
- `refactor`: A code change that neither fixes a bug nor adds a feature.
- `perf`: A code change that improves performance.
- `test`: Adding missing tests or correcting existing tests.
- `chore`: Changes to build process or auxiliary tools (librarian, etc).

**Example**:
```text
feat(api): implemented transaction list endpoint with pagination

Added Cursor based pagination for /v1/transactions.
Query performance optimized with composite index.

Closes SC-202
```

---

## 2. Go Style Guide (Uber Standard)
Based on [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

### 2.1 Interface
- Define interfaces where they are **used**, not where they are implemented.
- Keep interfaces small (1-2 methods).

### 2.2 Error Handling
- **Never ignore errors**. Handle them or return them.
- Use `fmt.Errorf("context: %w", err)` to wrap errors.
- Don't use `panic` in production code (only in `main` or `init`).

### 2.3 Concurrency
- **Avoid** `context.Background()` in library code. Pass context explicitly.
- Use `errgroup` or `waitgroup` to manage goroutines.
- Always handle channel closure; avoid data races (use `-race` in tests).

---

## 3. Directory Structure (Hexagonal)
```text
/cmd
  /api          # Main entrypoint
  /migrate      # DB migration tool
/internal
  /domain       # Enterprise Business Rules (Entities)
  /platform     # Adapters (Postgres, Redis, Kafka)
  /server       # HTTP Handlers & Routes
  /service      # Application Logic (Use Cases)
/pkg            # Public Libraries (Logger, Config)
/migrations     # SQL files
```

---

## 4. Testing Strategy
- **Unit Tests**: Table-driven tests. Mock external dependencies.
- **Integration Tests**: Verify DB interactions (Testcontainers).
- **Naming**: `TestEntity_Scenario_Outcome`.

---

## 5. Security Checklist
- No secrets in code (use ENV).
- Validate all inputs (Length, Type, Range).
- Sanitize logic before SQL query (Use Placeholders `$1`).
- Rate Limit public endpoints.
