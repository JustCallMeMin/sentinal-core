#!/bin/bash
# Create Epic Issues for Sentinal Core

# Epic 1: Project Foundation
gh issue create \
  --title "[EPIC 1] Project Foundation (SC-001 → SC-010)" \
  --body "**Goal**: Establish project scaffold, config, logging, and build system.

## Tasks
- [x] SC-001: Create project scaffold
- [ ] SC-002: Setup config loader (.env + validation)
- [ ] SC-003: Setup structured JSON logger with trace_id
- [ ] SC-004: Setup Makefile (run, test, build)
- [ ] SC-005: Add /healthz endpoint
- [ ] SC-006: Add /readyz endpoint (DB check)
- [ ] SC-007: Setup graceful shutdown
- [ ] SC-008: Setup Docker multi-stage build
- [ ] SC-009: Setup docker-compose (API + DB + Redis)
- [ ] SC-010: Add linter + pre-commit hook

**Gate A**: App boots via Docker, Health endpoints working, Lint passes" \
  --label "epic,foundation"

# Epic 2: Database & Migration
gh issue create \
  --title "[EPIC 2] Database & Migration (SC-011 → SC-020)" \
  --body "**Goal**: Production-ready database layer with migrations and testing.

## Tasks
- [ ] SC-011: Integrate pgxpool
- [ ] SC-012: Add DB timeout & max connection config
- [ ] SC-013: Integrate migration tool (golang-migrate)
- [ ] SC-014: Test forward migration locally
- [ ] SC-015: Test rollback migration
- [ ] SC-016: Add testcontainer integration test
- [ ] SC-017: Seed initial tenant + admin user
- [ ] SC-018: Add DB index validation test
- [ ] SC-019: Add transaction wrapper helper
- [ ] SC-020: Add query timeout enforcement

**Gate B**: Migration reversible, Integration test green, Query timeout enforced" \
  --label "epic,database"

# Epic 3: Auth & RBAC
gh issue create \
  --title "[EPIC 3] Auth & RBAC (SC-021 → SC-030)" \
  --body "**Goal**: Secure authentication and role-based access control.

## Tasks
- [ ] SC-021: Implement password hashing (bcrypt)
- [ ] SC-022: Implement login endpoint
- [ ] SC-023: Implement JWT issuance
- [ ] SC-024: Implement JWT middleware
- [ ] SC-025: Implement role permission matrix
- [ ] SC-026: Enforce permission on admin endpoints
- [ ] SC-027: Implement account lockout logic
- [ ] SC-028: Implement MFA verification flow
- [ ] SC-029: Add API key middleware (tenant)
- [ ] SC-030: Add API key revocation + last_used_at update

**Gate C**: Expired token rejected, Locked account blocked, Permission enforcement verified" \
  --label "epic,security"

# Epic 4: Transaction Read API
gh issue create \
  --title "[EPIC 4] Transaction Read API (SC-031 → SC-038)" \
  --body "**Goal**: High-performance transaction query API with pagination.

## Tasks
- [ ] SC-031: Implement POST /transactions (ingest)
- [ ] SC-032: Implement idempotency via correlation_id
- [ ] SC-033: Implement GET /transactions (cursor pagination)
- [ ] SC-034: Add index on (tenant_id, occurred_at DESC)
- [ ] SC-035: Implement GET /transactions/:id
- [ ] SC-036: Implement filtering by outcome
- [ ] SC-037: Add integration test with 50k records
- [ ] SC-038: Benchmark list query < 50ms on 100k rows

**Gate D**: No duplicate ingestion, Pagination stable, Query under latency budget" \
  --label "epic,api"

# Epic 5: Rule Engine
gh issue create \
  --title "[EPIC 5] Rule Engine (SC-039 → SC-046)" \
  --body "**Goal**: Deterministic fraud detection rule engine.

## Tasks
- [ ] SC-039: Define rule schema struct
- [ ] SC-040: Implement numeric comparator
- [ ] SC-041: Implement string comparator
- [ ] SC-042: Implement IN operator
- [ ] SC-043: Implement threshold rule evaluation
- [ ] SC-044: Add table-driven unit tests
- [ ] SC-045: Add edge-case tests
- [ ] SC-046: Benchmark rule evaluation < 5ms

**Gate E**: 100% branch coverage rule engine, <5ms evaluation" \
  --label "epic,ml"

# Epic 6: ML Bridge
gh issue create \
  --title "[EPIC 6] ML Bridge (SC-047 → SC-054)" \
  --body "**Goal**: Integrate ML scoring service with circuit breaker.

## Tasks
- [ ] SC-047: Define gRPC proto for predict
- [ ] SC-048: Implement ML mock server
- [ ] SC-049: Implement gRPC client
- [ ] SC-050: Add timeout handling (30ms)
- [ ] SC-051: Add retry policy (max 1)
- [ ] SC-052: Implement circuit breaker
- [ ] SC-053: Add fallback to rule-only mode
- [ ] SC-054: Integration test ML-down scenario

**Gate F**: ML down → API still returns, No request exceeds 120ms" \
  --label "epic,ml"

# Epic 7: Scoring Endpoint
gh issue create \
  --title "[EPIC 7] Scoring Endpoint (SC-055 → SC-061)" \
  --body "**Goal**: Real-time fraud scoring with <80ms P95 latency.

## Tasks
- [ ] SC-055: Implement POST /score flow orchestration
- [ ] SC-056: Persist decision async (mock Kafka)
- [ ] SC-057: Persist feature_snapshot
- [ ] SC-058: Emit label event stub
- [ ] SC-059: Add latency instrumentation
- [ ] SC-060: Benchmark P95 < 80ms
- [ ] SC-061: Load test script (k6)

**Gate G**: P95 < 80ms, P99 < 120ms, No DB blocking before response" \
  --label "epic,api"

# Epic 8: Review Workflow
gh issue create \
  --title "[EPIC 8] Review Workflow (SC-062 → SC-068)" \
  --body "**Goal**: Manual review queue with concurrency control.

## Tasks
- [ ] SC-062: Implement GET /reviews?status=pending
- [ ] SC-063: Implement claim review (SELECT FOR UPDATE)
- [ ] SC-064: Prevent double-claim concurrency test
- [ ] SC-065: Implement POST /reviews/:id/resolve
- [ ] SC-066: Persist label from review
- [ ] SC-067: Add SLA timestamp tracking
- [ ] SC-068: Integration test review lifecycle

**Gate H**: No race condition, Label correctly stored" \
  --label "epic,api"

# Epic 9: Metrics & Dashboard
gh issue create \
  --title "[EPIC 9] Metrics & Dashboard (SC-069 → SC-074)" \
  --body "**Goal**: Analytics API for dashboard visualization.

## Tasks
- [ ] SC-069: Implement daily_metrics aggregation job
- [ ] SC-070: Implement GET /metrics endpoint
- [ ] SC-071: Add index for metrics query
- [ ] SC-072: Add fraud_rate calculation validation test
- [ ] SC-073: Add audit log on policy change
- [ ] SC-074: Add audit query endpoint

**Gate I**: Metrics endpoint < 20ms, Audit trail retrievable" \
  --label "epic,analytics"

# Epic 10: Production Hardening
gh issue create \
  --title "[EPIC 10] Production Hardening (SC-075 → SC-078)" \
  --body "**Goal**: Observability, rate limiting, and CI/CD.

## Tasks
- [ ] SC-075: Add Prometheus metrics
- [ ] SC-076: Add OpenTelemetry tracing
- [ ] SC-077: Add Redis rate limiter (token bucket)
- [ ] SC-078: Setup GitHub Actions CI (test + build)

**Gate J**: CI green on PR, Rate limit returns 429, Trace spans visible" \
  --label "epic,devops"

echo "✅ Created 10 Epic Issues"
