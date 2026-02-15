# System Architecture: Enterprise Fraud Platform (v2)

## Design Principles
- **Non-blocking Decision Path**: Results are returned before long-running I/O.
- **Rule Fast-path**: Lightweight in-memory heuristics before ML.
- **Unified Feature Registry (Feast)**: Eliminate training-serving skew via shared code paths.
- **Precompute Feature Bundles**: Aggregated feature vectors for single-fetch efficiency.
- **Deterministic Replay**: Immutable audit storage (S3 + Hash chain) and replay tools.
- **Cost-aware Tuning**: Optimize thresholds based on financial loss vs. fraud risk.
- **Graceful Degradation**: Circuit breakers for all external service dependencies.

## Conceptual Layers

### 1. Real-time Decision Layer
- **API Gateway**: Tenant resolution, authentication, and rate limiting.
- **Rule Engine**: In-memory execution of blacklists and velocity caps.
- **Scoring Service**: Coordinates feature retrieval and ML inference.
- **Circuit Breaker**: Fail-safe mechanisms for Redis and ML service outages.

### 2. Data & Feature Layer
- **Online Feature Store (Feast/Redis)**: Low-latency storage for precomputed feature bundles.
- **Offline Data Store (PostgreSQL)**: Partitioned by tenant for high-performance retrieval.
- **Event Bus (Kafka)**: Partitioned per tenant for isolated replay and auditing.

### 3. ML Lifecycle Layer
- **Inference Service**: Stateless ONNX runner (POST /predict).
- **Training Pipeline**: Batch processing with automated label collection.
- **Replay Engine**: Tool for evaluating model performance on historical data.

## Data Flow
1. **Request**: Transaction sent from Merchant.
2. **Fast-path**: Rule engine checks for immediate block/allow (Circuit-breaker protected).
3. **Hydrate**: Fetch **Precomputed Bundle** from Feast Online Store.
4. **Predict**: ML service provides probability score.
5. **Decide**: Policy engine applies merchant-specific **Cost-aware** thresholds.
6. **Response**: Return decision to merchant (<80ms).
7. **Emit**: Send scoring event to Kafka for async auditing, SHAP explainability, and replay.
