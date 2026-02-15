# Sentinal Core - Architecture Decision Records (ADR)

> **Context**: Real-Time Fraud Detection SaaS (Elite Tier).
> **Goal**: 50ms Latency for Scoring.

---

## 1. Primary Database: PostgreSQL (v16+)
**Decision**: Use PostgreSQL with **Declarative List Partitioning** (`tenant_id`).
**Reason**:
- **Pro**: Enterprise Feature (Scale to Billions of rows per Tenant).
- **Pro**: Data Locality (Drop Tenant = Drop Table).
- **Con**: Complex migration (Solved by `golang-migrate` & `bootstrap_data.py`).
**Status**: ACCEPTED.

---

## 2. API Framework: Go Fiber (v2)
**Decision**: Use `gofiber/fiber` (FastHTTP based).
**Reason**:
- **Pro**: Zero Memory Allocation (Performance).
- **Pro**: Express-like Middleware API (Easy for Node devs).
- **Pro**: Built-in Rate Limiter & Compression.
**Status**: ACCEPTED.

---

## 3. Asynchronous Messaging: Redis Streams (+ Kafka option)
**Decision**: Start with Redis Streams (Simpler), migrate to Kafka later.
**Reason**:
- **Pro**: Low latency (< 1ms).
- **Pro**: Simple consumer groups.
- **Con**: Data durability relies on AOF persistence.
**Status**: ACCEPTED (Phase 1-3).

---

## 4. ML Integration: gRPC (Go -> Python) with ONNX fallback
**Decision**: Microservice Architecture.
**Reason**:
- **Pro**: Decouple Go (Fast API) from Python (ML Libraries).
- **Pro**: Strongly typed contracts (Protobuf).
- **Alternative**: Embed ONNX Runtime in Go (Risk: CGO Complexity).
**Status**: ACCEPTED.

---

## 5. Deployment Strategy: Docker Multi-Stage Build and Distroless
**Decision**: Use `distroless/static` for final image.
**Reason**:
- **Pro**: Smallest image size (< 20MB).
- **Pro**: Security (No shell exploits).
**Status**: ACCEPTED.

---

## 6. Authentication: JWT + API Key (Redis Cache)
**Decision**: Dual Auth Mechanism.
- **Admin**: JWT (Short-lived 15m + Refresh Token).
- **API**: Long-lived API Key (Cached in Redis).
**Reason**: Separate concerns for Human vs Machine.
**Status**: ACCEPTED.
