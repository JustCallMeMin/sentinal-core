# Business Workflows & Process Flows

## 1. Real-time Transaction Scoring Flow
1. **Initiation**: Merchant system calls `POST /v1/score` with transaction details.
2. **Identification**: Gateway identifies `tenant_id`, checks session quotas, and caches `correlation_id`.
3. **Screening**: 
    - Rule Engine checks if the IP or User is in the global/tenant blacklist (Circuit-breaker protected).
    - If match: Return **Block** (3ms).
4. **Hydration**: If no early match, scoring service fetches **Precomputed Bundle** from Feast Online Store.
5. **Inference**: ML service calculates Fraud Probability Score using industry-specific segment models.
6. **Decisioning**: Policy Engine checks score against **Cost-aware thresholds**:
    - Based on Financial Loss vs. Fraud Probability.
7. **Response**: Return decision to merchant (<80ms).
8. **Audit**: Send `TransactionScoredEvent` to Kafka for async auditing, SHAP generation, and analytics.

## 2. SHAP & Manual Review Flow
1. **Trigger**: Transaction is marked as "Challenge".
2. **Async Processor**: Kafka consumer triggers SHAP explainability calculation.
3. **Queue**: Request appears in the Reviewer's Task Queue with a visual SHAP importance graph.
4. **Decision**: Reviewer analyzes `Feature Snapshot` and `SHAP Insights` and chooses **Approve** or **Confirm Fraud**.
5. **Feedback**: System emits `LabelUpdatedEvent`.
6. **Retrain**: Batch job picks up new labels to retrain the ML model every N hours.

## 3. Merchant Onboarding (Future)
1. **Sign-up**: Merchant registers and receives a unique `tenant_id`.
2. **Key Creation**: Merchant generates API Keys.
3. **Policy Tuning**: Merchant sets their own thresholds for Approve/Challenge/Block.
