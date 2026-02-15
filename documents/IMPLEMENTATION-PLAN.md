# Sentinal Core - Execution Kanban (v7.0)

**Total Tasks**: 78
**Estimate**: ~45-55 dev days (Solo)
**Methodology**: Atomic Kanban (Task <= 1 day)

---

## 🛑 DEFINITION OF DONE (DoD)
Every task is considered DONE only when:
1.  Code committed to `main` (no long-lived branches).
2.  Unit Tests added & passed (`go test -race`).
3.  Integration Tests passed (if DB involved).
4.  Linter check passed (`golangci-lint` - no new issues).
5.  Logs validated (JSON format, Trace ID present).

---

## 🏗️ EPIC 1: Project Foundation (SC-001 → SC-010)
- [ ] **SC-001**: Create project scaffold (`cmd/api`, `internal`, `pkg`) – `0.5d` [Dep: None]
- [ ] **SC-002**: Setup config loader (`.env` + validation) – `0.5d` [Dep: SC-001]
- [ ] **SC-003**: Setup structured JSON logger with `trace_id` – `0.5d` [Dep: SC-002]
- [ ] **SC-004**: Setup `Makefile` (run, test, build) – `0.5d` [Dep: SC-001]
- [ ] **SC-005**: Add `/healthz` endpoint – `0.5d` [Dep: SC-001]
- [ ] **SC-006**: Add `/readyz` endpoint (DB check) – `0.5d` [Dep: SC-005]
- [ ] **SC-007**: Setup graceful shutdown – `0.5d` [Dep: SC-003]
- [ ] **SC-008**: Setup Docker multi-stage build – `0.5d` [Dep: SC-004]
- [ ] **SC-009**: Setup `docker-compose` (API + DB + Redis) – `0.5d` [Dep: SC-008]
- [ ] **SC-010**: Add linter + pre-commit hook – `0.5d` [Dep: SC-004]

**🚦 Gate A**:
- [ ] App boots via Docker
- [ ] Health endpoints working
- [ ] Lint passes

---

## 🗄️ EPIC 2: Database & Migration (SC-011 → SC-020)
- [ ] **SC-011**: Integrate `pgxpool` – `0.5d` [Dep: SC-002]
- [ ] **SC-012**: Add DB timeout & max connection config – `0.5d` [Dep: SC-011]
- [ ] **SC-013**: Integrate migration tool (`golang-migrate`) – `0.5d` [Dep: SC-011]
- [ ] **SC-014**: Test forward migration locally – `0.5d` [Dep: SC-013]
- [ ] **SC-015**: Test rollback migration – `0.5d` [Dep: SC-014]
- [ ] **SC-016**: Add `testcontainer` integration test – `1d` [Dep: SC-015]
- [ ] **SC-017**: Seed initial tenant + admin user – `0.5d` [Dep: SC-016]
- [ ] **SC-018**: Add DB index validation test – `0.5d` [Dep: SC-017]
- [x] **SC-019**: Add transaction wrapper helper – `0.5d` [Dep: SC-011]
- [x] **SC-020**: Add query timeout enforcement – `0.5d` [Dep: SC-012]

**🚦 Gate B**:
- [x] Migration reversible
- [x] Integration test green (Testcontainers)
- [x] Soft Delete verified

---

## 🔐 EPIC 3: Auth & RBAC (SC-021 → SC-030)
- [x] **SC-021**: Implement password hashing (`bcrypt`) – `0.5d` [Dep: None] ✅ 2026-02-15
- [x] **SC-022**: Implement login endpoint – `0.5d` [Dep: SC-021, SC-011] ✅ 2026-02-15
- [ ] **SC-023**: Implement JWT issuance – `0.5d` [Dep: SC-022]
- [ ] **SC-024**: Implement JWT middleware – `0.5d` [Dep: SC-023]
- [ ] **SC-025**: Implement role permission matrix – `1d` [Dep: None]
- [ ] **SC-026**: Enforce permission on admin endpoints – `0.5d` [Dep: SC-025]
- [ ] **SC-027**: Implement account lockout logic – `0.5d` [Dep: SC-022]
- [ ] **SC-028**: Implement MFA verification flow – `1d` [Dep: SC-022]
- [ ] **SC-029**: Add API key middleware (tenant) – `0.5d` [Dep: SC-011]
- [ ] **SC-030**: Add API key revocation + `last_used_at` update – `0.5d` [Dep: SC-029]

**🚦 Gate C**:
- [ ] Expired token rejected
- [ ] Locked account blocked
- [ ] Permission enforcement verified

---

## 📜 EPIC 4: Transaction Read API (SC-031 → SC-038)
- [x] **SC-031**: Implement `POST /transactions` (ingest) – `0.5d` [Dep: SC-011, SC-029] ✅ 2026-02-15
- [x] **SC-032**: Implement idempotency via `correlation_id` – `0.5d` [Dep: SC-031] ✅ 2026-02-15
- [ ] **SC-033**: Implement `GET /transactions` (cursor pagination) – `1d` [Dep: SC-031]
- [ ] **SC-034**: Add index on (`tenant_id`, `occurred_at` DESC) – `0.5d` [Dep: SC-033]
- [ ] **SC-035**: Implement `GET /transactions/:id` – `0.5d` [Dep: SC-033]
- [ ] **SC-036**: Implement filtering by outcome – `0.5d` [Dep: SC-033]
- [ ] **SC-037**: Add integration test with 50k records – `1d` [Dep: SC-033]
- [ ] **SC-038**: Benchmark list query < 50ms on 100k rows – `0.5d` [Dep: SC-037]

**🚦 Gate D**:
- [ ] **Dataset**: 100k Transactions seeded.
- [ ] **Load**: 200 RPS via k6.
- [ ] **HW**: Local Docker (2 CPU/4GB RAM).
- [ ] **Metric**: P95 Latency < 50ms.

---

## 🧠 EPIC 5: Rule Engine (SC-039 → SC-046)
- [ ] **SC-039**: Define rule schema struct – `0.5d` [Dep: None]
- [ ] **SC-040**: Implement numeric comparator – `0.5d` [Dep: SC-039]
- [ ] **SC-041**: Implement string comparator – `0.5d` [Dep: SC-039]
- [ ] **SC-042**: Implement IN operator – `0.5d` [Dep: SC-041]
- [ ] **SC-043**: Implement threshold rule evaluation – `0.5d` [Dep: SC-040]
- [ ] **SC-044**: Add table-driven unit tests – `0.5d` [Dep: SC-043]
- [ ] **SC-045**: Add edge-case tests – `0.5d` [Dep: SC-044]
- [ ] **SC-046**: Benchmark rule evaluation < 5ms – `0.5d` [Dep: SC-044]

**🚦 Gate E**:
- [ ] 100% branch coverage rule engine
- [ ] <5ms evaluation (100 complex rules)

---

## 🤖 EPIC 6: ML Bridge (SC-047 → SC-054)
- [ ] **SC-047**: Define gRPC proto for predict – `0.5d` [Dep: None]
- [ ] **SC-048**: Implement ML mock server – `0.5d` [Dep: SC-047]
- [ ] **SC-049**: Implement gRPC client – `0.5d` [Dep: SC-047]
- [ ] **SC-050**: Add timeout handling (30ms) – `0.5d` [Dep: SC-049]
- [ ] **SC-051**: Add retry policy (max 1) – `0.5d` [Dep: SC-050]
- [ ] **SC-052**: Implement circuit breaker – `1d` [Dep: SC-050]
- [ ] **SC-053**: Add fallback to rule-only mode – `0.5d` [Dep: SC-052]
- [ ] **SC-054**: Integration test ML-down scenario – `0.5d` [Dep: SC-053]

**🚦 Gate F**:
- [ ] ML Service Down -> API still returns 200 (Fallback).
- [ ] No request exceeds 120ms (Circuit Breaker verified).

---

## ⚡ EPIC 7: Scoring Endpoint (SC-055 → SC-061)
- [ ] **SC-055**: Implement `POST /score` flow orchestration – `1d` [Dep: SC-043, SC-049]
- [ ] **SC-056**: Persist decision async (mock Kafka) – `0.5d` [Dep: SC-055]
- [ ] **SC-057**: Persist `feature_snapshot` – `0.5d` [Dep: SC-055]
- [ ] **SC-058**: Emit label event stub – `0.5d` [Dep: SC-055]
- [ ] **SC-059**: Add latency instrumentation – `0.5d` [Dep: SC-055]
- [ ] **SC-060**: Benchmark P95 < 80ms – `1d` [Dep: SC-055]
- [ ] **SC-061**: Load test script (k6) – `0.5d` [Dep: SC-060]

**🚦 Gate G**:
- [ ] P95 < 80ms (200 RPS, 100k Rows).
- [ ] Async write queue not blocking HTTP response.

---

## 👁️ EPIC 8: Review Workflow (SC-062 → SC-068)
- [ ] **SC-062**: Implement `GET /reviews?status=pending` – `0.5d` [Dep: SC-011]
- [ ] **SC-063**: Implement claim review (SELECT FOR UPDATE) – `0.5d` [Dep: SC-062]
- [ ] **SC-064**: Prevent double-claim concurrency test – `0.5d` [Dep: SC-063]
- [ ] **SC-065**: Implement `POST /reviews/:id/resolve` – `0.5d` [Dep: SC-063]
- [ ] **SC-066**: Persist label from review – `0.5d` [Dep: SC-065]
- [ ] **SC-067**: Add SLA timestamp tracking – `0.5d` [Dep: SC-065]
- [ ] **SC-068**: Integration test review lifecycle – `0.5d` [Dep: SC-065]

**🚦 Gate H**:
- [ ] No race condition on Claim.
- [ ] Label created after Resolution.

---

## 📊 EPIC 9: Metrics & Dashboard (SC-069 → SC-074)
- [ ] **SC-069**: Implement `daily_metrics` aggregation job – `1d` [Dep: SC-011]
- [ ] **SC-070**: Implement `GET /metrics` endpoint – `0.5d` [Dep: SC-069]
- [ ] **SC-071**: Add index for metrics query – `0.5d` [Dep: SC-070]
- [ ] **SC-072**: Add `fraud_rate` calculation validation test – `0.5d` [Dep: SC-070]
- [ ] **SC-073**: Add audit log on policy change – `0.5d` [Dep: SC-039]
- [ ] **SC-074**: Add audit query endpoint – `0.5d` [Dep: SC-073]

**🚦 Gate I**:
- [ ] Metrics endpoint < 20ms.
- [ ] Audit trail works for Policy updates.

---

## 🛡️ EPIC 10: Production Hardening (SC-075 → SC-078)
- [ ] **SC-075**: Add Prometheus metrics – `0.5d` [Dep: SC-003]
- [ ] **SC-076**: Add OpenTelemetry tracing – `0.5d` [Dep: SC-003]
- [ ] **SC-077**: Add Redis rate limiter (token bucket) – `0.5d` [Dep: SC-002]
- [ ] **SC-078**: Setup GitHub Actions CI (test + build) – `1d` [Dep: None]

**🚦 Gate J**:
- [ ] CI green on PR.
- [ ] Rate limit returns 429.
- [ ] Trace spans visible.
