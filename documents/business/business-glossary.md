# Business Glossary & Definitions

| Term | Definition |
| --- | --- |
| **Tenant** | A unique merchant or customer using the platform. Data and models are isolated per tenant. |
| **Scoring** | The process of assigning a risk value to a transaction. |
| **Fast-path** | A lightweight rule-based evaluation that happens before complex ML inference. |
| **Challenge** | A status indicating a transaction needs manual human review (medium risk). |
| **Feature Parity**| The guarantee that the data used to train a model is identical to the data provided during real-time scoring. |
| **Training-Serving Skew** | A common bug where feature values differ between training and serving environments. |
| **Feature Bundle** | An aggregated object containing multiple feature keys, fetched in a single Redis I/O. |
| **Circuit Breaker** | A pattern used to detect failures and prevent them from cascading (e.g., if ML service is down). |
| **SHAP (Async)** | SHapley Additive exPlanations calculated asynchronously to explain AI decisions. |
| **Cost-aware Tuning** | Tuning model thresholds to minimize financial loss rather than just error rate. |
| **mTLS** | Mutual TLS for secure, authenticated communication between all internal microservices. |
| **Replay Engine** | A tool to run historical data through new models to evaluate performance uplift. |
