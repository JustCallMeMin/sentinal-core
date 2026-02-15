# Project Overview: Sentinal Core

Sentinal Core is a production-grade, multi-tenant Fraud Detection and Trust Scoring Platform designed for B2B SaaS. It provides real-time transaction scoring using a hybrid approach of high-performance heuristics (Rule Engine) and Machine Learning models.

## Goals
- **Real-time Performance**: Sub-80ms end-to-end decision latency.
- **Enterprise Reliability**: Asynchronous persistence and high availability.
- **Feature Parity**: Guarantee consistency between training and serving data.
- **Multi-tenancy**: Strict isolation between merchant data and models.

## Tech Stack
- **API Gateway/Ingestion**: Go (high concurrency, low latency).
- **Feature Store**: **Feast** (Centralized feature registry and low-latency serving via Redis Cluster).
- **Inference Service**: Python (FastAPI + ONNX Runtime).
- **Security & Communication**: **mTLS** between services, HMAC-signed requests.
- **Data Layers**:
  - Redis Cluster (Online storage for precomputed feature bundles).
  - PostgreSQL (Partitioned by `tenant_id` for enterprise scale).
  - Apache Kafka (Partitioned per tenant for replay and audit).
- **Observability**: Prometheus, OpenTelemetry, and **SLO-based Dashboards**.

## Getting Started
(Initial documentation phase. Standard Go/Python setup instructions will be added as implementation begins.)
