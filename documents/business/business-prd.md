# Business Product Requirements Document (PRD)

## Product Vision
Sentinal Core aims to be the most reliable and transparent Fraud Detection Platform for modern B2B SaaS companies, offering sub-100ms risk scoring with a human-in-the-loop review system.

## Stakeholders
- **Merchants**: The primary users who integrate the API to protect their transactions.
- **Risk Analysts (Reviewers)**: Human operators who handle challenged/ambiguous transactions.
- **Developers**: Integration engineers who interact with the API Gateway and SDKs.
- **Compliance Officers**: Need logs and feature snapshots for regulatory auditing.

## Target Audience
- Mid-to-large scale B2B platforms requiring isolated tenant data.
- FinTech gateways needing **Sector-specific Models** (e.g., Luxury Goods vs. Digital Content).
- Enterprise platforms requiring strict auditability and deterministic replay.

## Success Metrics
- **False Positive Cost (FPC)**: Total financial loss attributed to churn/bad UX from legitimate blocks.
- **Estimated Fraud Savings**: Total fraud amount blocked vs. loss from False Negatives.
- **Inference Latency (P99)**: < 30ms (In-memory/ONNX path).
- **Service SLOs**: Aggregated P99 < 80ms for the entire scoring request.
- **Review SLA**: 95% of challenged transactions reviewed within 4 hours.
