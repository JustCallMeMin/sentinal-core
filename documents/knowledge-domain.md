# Domain Model & Data Standards

## Core Entities

### Tenant (Root Entity)
- `id`: UUID (Primary Key)
- `name`: String
- `industry_segment`: String (e.g. Retail, Finance)
- `api_key_hash`: String (Hashed)
- `status`: Enum (Active, Suspended)
- `deleted_at`: Timestamp (Nullable - Soft Delete support)

### Transaction
- `id`: UUID (Primary Key)
- `tenant_id`: UUID (Isolation Key)
- `user_id`: String (Reference to merchant's user)
- `amount`: Decimal
- `currency`: ISO String
- `device_id`: String
- `ip_address`: String
- `payload`: JSONB (Raw transaction details)

### Decision
- `transaction_id`: FK -> Transaction
- `score`: Float (0.0 to 1.0)
- `decision`: Enum (Approve, Challenge, Block)
- `model_version`: String
- `policy_version`: String

### Feature Snapshot (Mandatory)
- `transaction_id`: FK -> Transaction
- `features_json`: Map of full feature vector at the time of scoring.
- **Requirement**: Must be stored for every inference to enable deterministic replay and audit compliance.

### Label (Ground Truth)
- `id`: UUID (Primary Key)
- `transaction_id`: FK -> Transaction
- `tenant_id`: UUID (Partition Key)
- `source`: Enum (Manual Review, Chargeback, Provider)
- `value`: Enum (Fraud, Legit)
- **Importance**: Decouples the final "truth" from the initial AI prediction.

### SaaS Identity & Security (Elite Tier v3.5)
- **User**: Includes `mfa_secret` and `status` lifecycle (Active/Locked).
- **RBAC (Normalized)**: Roles and Permissions are now M:N relation (`roles`, `permissions`, `role_permissions`). JSONB deprecated for strict governance.
- **Identity Ops**: Supporting Password Reset/Email Verify via `identity_verification_tokens`.

### Mobile Infrastructure
- **Device Registry**: `device_tokens` table tracks FCM/APNS tokens per user-device pair.
- **Notification Center**: Partitioned `notifications` table for scalable alerting.

### Infrastructure & Scaling
- **Key Rotation**: `tenant_api_keys` includes `revoked_at`, `scopes` (JSONB), and rotation schedule.
- **Indexing Strategy**: Mandatory indexes on (`tenant_id`, `status`) and (`tenant_id`, `occurred_at` DESC).

### SaaS Governance
- **Billing & Usage**: Tracking API usage quotas and compute consumption for tier enforcement.
- **Enhanced Audit**: Explicit logging of High-Privilege actions (Policy changes, Model switches).

## Business Rules
- **Multi-tenancy**: Every query must include a `tenant_id`.
- **Soft-delete Policy**: Management entities are never hard-deleted; `deleted_at` timestamp is used.
- **Security-First**: Rate Limiting applied at Gateway and Dashboard APIs. MFA is mandatory for Admin roles.
- **Pagination**: All list APIs must implement cursor-based or offset-based pagination with strict size limits.

## Database Schema (PostgreSQL)
- **Transactional**: `transactions`, `decisions`, `feature_snapshots`, `labels`, `review_tasks`, `daily_metrics`.
- **SaaS Ops**: `users`, `roles`, `permissions`, `billing_tiers`, `usage_logs`.
