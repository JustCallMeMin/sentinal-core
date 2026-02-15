# Sentinal Core - Execution Kanban (v7.1)

**Total Tasks**: 80
**Estimate**: ~45-55 dev days (Solo)
**Methodology**: Atomic Kanban (Task <= 1 day)

---

## 🏗️ EPIC 1: Project Foundation (SC-001 → SC-010)
- [x] **SC-001**: Create project scaffold ✅ 2026-02-15
- [x] **SC-002**: Setup config loader ✅ 2026-02-15
- [x] **SC-003**: Setup structured JSON logger ✅ 2026-02-15
- [x] **SC-004**: Setup Makefile ✅ 2026-02-15
- [x] **SC-005**: Add /healthz endpoint ✅ 2026-02-15
- [x] **SC-006**: Add /readyz endpoint ✅ 2026-02-15
- [x] **SC-007**: Setup graceful shutdown ✅ 2026-02-15
- [x] **SC-008**: Setup Docker multi-stage build ✅ 2026-02-15
- [x] **SC-009**: Setup docker-compose ✅ 2026-02-15
- [x] **SC-010**: Add linter + pre-commit hook ✅ 2026-02-15

---

## 🗄️ EPIC 2: Database & Migration (SC-011 → SC-022)
- [x] **SC-011**: Integrate pgxpool ✅ 2026-02-15
- [x] **SC-012**: Add DB timeout & config ✅ 2026-02-15
- [x] **SC-013**: Integrate migration tool ✅ 2026-02-15
- [x] **SC-014**: Test forward migration ✅ 2026-02-15
- [x] **SC-015**: Test rollback migration ✅ 2026-02-15
- [x] **SC-016**: Add testcontainer integration test ✅ 2026-02-15
- [x] **SC-017**: Seed initial tenant + admin user ✅ 2026-02-15
- [x] **SC-018**: Add DB index validation test ✅ 2026-02-15
- [x] **SC-019**: Add transaction wrapper helper ✅ 2026-02-15
- [x] **SC-020**: Add query timeout enforcement ✅ 2026-02-15
- [x] **SC-021**: Handle DB errors and mapping ✅ 2026-02-15
- [x] **SC-022**: Configure connection pool for high load ✅ 2026-02-15

---

## 🔐 EPIC 3: Auth & RBAC (SC-023 → SC-032)
- [x] **SC-023**: Implement password hashing (bcrypt) ✅ 2026-02-15
- [x] **SC-024**: Implement login endpoint ✅ 2026-02-15
- [x] **SC-025**: Implement JWT issuance ✅ 2026-02-15
- [x] **SC-026**: Implement JWT Auth middleware ✅ 2026-02-15
- [x] **SC-027**: Implement role permission matrix (Hardened with JTI) ✅ 2026-02-15
- [x] **SC-028**: Enforce permission (Hardened with Audit Logs) ✅ 2026-02-15
- [ ] **SC-029**: Implement account lockout logic – `0.5d`
- [ ] **SC-030**: Implement MFA verification flow – `1d`
- [ ] **SC-031**: Add API key middleware (tenant) – `0.5d`
- [ ] **SC-032**: Add API key revocation + last_used_at update – `0.5d`

---

## 📜 EPIC 4: Transaction API (SC-033 → SC-040)
- [x] **SC-033**: Implement POST /transactions (ingest) ✅ 2026-02-15
- [x] **SC-034**: Implement idempotency via correlation_id ✅ 2026-02-15
- [ ] **SC-035**: Implement GET /transactions (cursor pagination) – `1d`
- [ ] **SC-036**: Add index on (tenant_id, occurred_at DESC) – `0.5d`
- [ ] **SC-037**: Implement GET /transactions/:id – `0.5d`
- [ ] **SC-038**: Implement filtering by outcome – `0.5d`
- [ ] **SC-039**: Add integration test with 50k records – `1d`
- [ ] **SC-040**: Benchmark list query < 50ms on 100k rows – `0.5d`

---

## 🧠 EPIC 5: Rule Engine (SC-041 → SC-048)
- [ ] **SC-041**: Define rule schema struct – `0.5d`
- [ ] **SC-042**: Implement numeric comparator – `0.5d`
- [ ] **SC-043**: Implement string comparator – `0.5d`
- [ ] **SC-044**: Implement IN operator – `0.5d`
- [ ] **SC-045**: Implement threshold rule evaluation – `0.5d`
- [ ] **SC-046**: Add table-driven unit tests – `0.5d`
- [ ] **SC-047**: Add edge-case tests – `0.5d`
- [ ] **SC-048**: Benchmark rule evaluation < 5ms – `0.5d`

---

## 🤖 EPIC 6: ML Bridge (SC-049 → SC-056)
- [ ] **SC-049**: Define gRPC proto for predict – `0.5d`
- [ ] **SC-050**: Implement ML mock server – `0.5d`
- [ ] **SC-051**: Implement gRPC client – `0.5d`
- [ ] **SC-052**: Add timeout handling (30ms) – `0.5d`
- [ ] **SC-053**: Add retry policy (max 1) – `0.5d`
- [ ] **SC-054**: Implement circuit breaker – `1d`
- [ ] **SC-055**: Add fallback to rule-only mode – `0.5d`
- [ ] **SC-056**: Integration test ML-down scenario – `0.5d`

---

## ⚡ EPIC 7: Scoring Endpoint (SC-057 → SC-063)
- [ ] **SC-057**: Implement POST /score flow orchestration – `1d`
- [ ] **SC-058**: Persist decision async (mock Kafka) – `0.5d`
- [ ] **SC-059**: Persist feature_snapshot – `0.5d`
- [ ] **SC-060**: Emit label event stub – `0.5d`
- [ ] **SC-061**: Add latency instrumentation – `0.5d`
- [ ] **SC-062**: Benchmark P95 < 80ms – `1d`
- [ ] **SC-063**: Load test script (k6) – `0.5d`

---

## 👁️ EPIC 8: Review Workflow (SC-064 → SC-070)
- [ ] **SC-064**: Implement GET /reviews?status=pending – `0.5d`
- [ ] **SC-065**: Implement claim review (SELECT FOR UPDATE) – `0.5d`
- [ ] **SC-066**: Prevent double-claim concurrency test – `0.5d`
- [ ] **SC-067**: Implement POST /reviews/:id/resolve – `0.5d`
- [ ] **SC-068**: Persist label from review – `0.5d`
- [ ] **SC-069**: Add SLA timestamp tracking – `0.5d`
- [ ] **SC-070**: Integration test review lifecycle – `0.5d`

---

## 📊 EPIC 9: Metrics & Dashboard (SC-071 → SC-076)
- [ ] **SC-071**: Implement daily_metrics aggregation job – `1d`
- [ ] **SC-072**: Implement GET /metrics endpoint – `0.5d`
- [ ] **SC-073**: Add index for metrics query – `0.5d`
- [ ] **SC-074**: Add fraud_rate calculation validation test – `0.5d`
- [ ] **SC-075**: Add audit log on policy change – `0.5d`
- [ ] **SC-076**: Add audit query endpoint – `0.5d`

---

## 🛡️ EPIC 10: Production Hardening (SC-077 → SC-080)
- [ ] **SC-077**: Add Prometheus metrics – `0.5d`
- [ ] **SC-078**: Add OpenTelemetry tracing – `0.5d`
- [ ] **SC-079**: Add Redis rate limiter (token bucket) – `0.5d`
- [ ] **SC-080**: Setup GitHub Actions CI – `1d`
