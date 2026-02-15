# Feature Specifications

## Core Feature Set (MVP)

### 1. Multi-Tenant API Ingestion
- **Description**: Secure endpoint for merchants to post transaction events.
- **Key Capability**: Tenant-aware rate limiting and schema validation.
- **Value**: Secure and isolated entry point for data.

### 2. Hybrid Scoring Engine
- **Description**: Combined Heuristic (Rules) and ML (ONNX) scoring.
- **Key Capability**: "Fast-path" early exit for certain-fraud/certain-pass cases.
- **Value**: Lower latency and optimized compute costs.

### 3. Unified Feature Registry
- **Description**: Single YAML-based definition for all ML features.
- **Key Capability**: Automatic parity between training pipelines and real-time serving.
- **Value**: Eliminates training-serving skew bugs.

### 4. Enterprise Replay Engine
- **Description**: Tool for batch re-evaluation of historical transactions.
- **Key Capability**: Compare model v1 vs. v2 performance using actual past data.
- **Value**: Demonstrates uplift and ROI before deploying new models.

### 5. Async Explainability (SHAP)
- **Description**: Detailed breakdown of feature importance per transaction.
- **Key Capability**: Visualizing why the AI flagged a specific event.
- **Value**: Transparency for manual reviewers and compliance auditors.

### 6. Audit & Compliance Vault
- **Description**: Immutable decision log with point-in-time feature snapshots and cryptographic hash chains.
- **Key Capability**: Guaranteed integrity of risk decisions.
- **Value**: Regulatory readiness for elite fintech environments.
