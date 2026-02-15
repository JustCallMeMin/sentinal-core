# Source Code Base Structure

## Directory Layout (Proposed)

```text
sentinal-core/
├── cmd/
│   ├── api/            # Go: Entry for Scoring API
│   ├── worker/         # Go: Async consumers (Audit, Sync, SHAP)
│   └── replay/         # Go: Historical replay & evaluation tool
├── internal/           # Go: Private application logic
│   ├── platform/       # Core infrastructure (Feast, Redis, Kafka, DB)
│   ├── rules/          # Heuristic logic engine
│   └── service/        # Orchestration (Inference, Feature Hydration)
├── ml-service/         # Python: FastAPI + ONNX Inference
├── feature-registry/   # Feast: Feature definitions and ML repo
├── documents/          # Project documentation
└── scripts/            # Deployment and CI/CD tools
```

## Key Modules
- **Rule Engine**: Written in Go for raw speed and memory efficiency.
- **Feast Registry**: Centralized source of truth for both online and offline feature paths.
- **Replay Tool**: Component for batch re-evaluation of historical transactions against new models.
- **Inference Wrapper**: Python thin-client to expose ONNX models via unified REST contract.

## Data Strategy (Bootstrapping)
- **Source**: IEEE-CIS Fraud Detection Dataset (Kaggle).
- **Location**: `.reports/ieee-fraud-detection/`.
- **Mapping Strategy**: 
    - `TransactionID` -> `correlation_id` (Unique ID).
    - `isFraud` -> `labels` table (0=Legit, 1=Fraud).
    - `TransactionDT` -> `occurred_at` (Offset from a base timestamp).
    - `Vxxx` (Vesta Features) -> `feature_snapshots.features_json`.
    - `card1` - `card6` -> `transactions` (payment_method, bin_country, etc.).
- **Goal**: Use this dataset to train the initial "Cold Start" model (XGBoost) before live traffic.
