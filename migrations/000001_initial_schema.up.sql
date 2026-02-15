-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For fuzzy search on names/emails

-- ==========================================
-- 0. UTILITY FUNCTIONS & TRIGGERS
-- ==========================================
-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ==========================================
-- 1. Identity & Tenant Layer (Core)
-- ==========================================
CREATE TABLE tenants (
    tenant_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    industry_segment VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended')),
    config JSONB DEFAULT '{}', -- Flexible config (thresholds, timeouts)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TRIGGER update_tenants_modtime BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- HARDENED API KEYS LIFECYCLE
CREATE TABLE tenant_api_keys (
    key_hash VARCHAR(64) PRIMARY KEY, -- SHA256
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    key_prefix VARCHAR(10) NOT NULL, -- To show "sk_live_1234..."
    key_name VARCHAR(100),
    scopes JSONB DEFAULT '["score:write"]', 
    
    -- Lifecycle & Rotation
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revoked_reason VARCHAR(255),
    rotation_scheduled_at TIMESTAMPTZ, -- For automated rotation policy
    
    created_by UUID, -- User ID who created this key
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_api_keys_tenant_active ON tenant_api_keys(tenant_id) WHERE revoked_at IS NULL;

-- ==========================================
-- 2. SaaS User & Normalized RBAC (Operations)
-- ==========================================
CREATE TABLE permissions (
    permission_id SERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL, -- e.g., "policy:edit", "report:view"
    description TEXT
);

CREATE TABLE roles (
    role_id SERIAL PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(tenant_id) ON DELETE CASCADE, -- Null means Global Role
    name VARCHAR(50) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE, -- Cannot be deleted
    UNIQUE(tenant_id, name)
);

CREATE TABLE role_permissions (
    role_id INT REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_id INT REFERENCES permissions(permission_id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id INT REFERENCES roles(role_id),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'locked')),
    failed_login_attempts INT DEFAULT 0,
    mfa_secret VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ, -- Soft delete
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_email_trgm ON users USING GIN (email gin_trgm_ops); -- Fast search
CREATE TRIGGER update_users_modtime BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Self-service Flow Tokens
CREATE TABLE identity_verification_tokens (
    token_hash VARCHAR(64) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('password_reset', 'email_verify')),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ==========================================
-- 3. Mobile & Notifications Support (Hardened)
-- ==========================================
CREATE TABLE device_tokens (
    device_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    fcm_token TEXT NOT NULL,
    platform VARCHAR(20) CHECK (platform IN ('ios', 'android', 'web')),
    is_active BOOLEAN DEFAULT TRUE,
    last_active_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_device_tokens_user ON device_tokens(user_id) WHERE is_active = TRUE;

CREATE TABLE notifications (
    notification_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL, -- Partition Key
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    type VARCHAR(50) NOT NULL, -- e.g., "review_assigned", "alert_threshold"
    priority VARCHAR(10) DEFAULT 'normal' CHECK (priority IN ('normal', 'high', 'critical')),
    is_read BOOLEAN DEFAULT FALSE,
    payload JSONB, -- Deep link data
    created_at TIMESTAMPTZ DEFAULT NOW(),
    read_at TIMESTAMPTZ,
    
    PRIMARY KEY (notification_id, tenant_id)
) PARTITION BY LIST (tenant_id);

CREATE INDEX idx_notifications_user_unread ON notifications(tenant_id, user_id) WHERE is_read = FALSE;

-- ==========================================
-- 4. Transaction Layer (Partitioned)
-- ==========================================
CREATE TABLE transactions (
    transaction_id UUID NOT NULL,
    tenant_id UUID NOT NULL, -- Partition Key
    correlation_id VARCHAR(128) NOT NULL,
    external_user_id VARCHAR(128),
    amount DECIMAL(18, 2) NOT NULL,
    currency CHAR(3) NOT NULL,
    ip_address INET,
    device_fingerprint VARCHAR(255),
    payment_method VARCHAR(50),
    bin_country CHAR(2),
    occurred_at TIMESTAMPTZ NOT NULL,
    payload JSONB, 
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (transaction_id, tenant_id),
    UNIQUE (tenant_id, correlation_id)
) PARTITION BY LIST (tenant_id);

-- Performance Indexes
CREATE INDEX idx_transactions_tenant_occurred ON transactions (tenant_id, occurred_at DESC);
CREATE INDEX idx_transactions_user_search ON transactions (tenant_id, external_user_id);
CREATE INDEX idx_transactions_ip ON transactions (tenant_id, ip_address);
CREATE INDEX idx_transactions_device ON transactions (tenant_id, device_fingerprint);

-- ==========================================
-- 5. Decision & Feature Layer (Partitioned)
-- ==========================================
CREATE TABLE decisions (
    transaction_id UUID NOT NULL,
    tenant_id UUID NOT NULL, -- Partition Key
    model_version_id UUID,   
    policy_id UUID,          
    score FLOAT NOT NULL,
    outcome VARCHAR(20) NOT NULL CHECK (outcome IN ('approve', 'challenge', 'block')),
    reason VARCHAR(255), -- Explanation code e.g. 'velocity_high'
    decided_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (transaction_id, tenant_id),
    FOREIGN KEY (transaction_id, tenant_id) REFERENCES transactions(transaction_id, tenant_id)
) PARTITION BY LIST (tenant_id);

-- Covering index for Dashboard Charts (Score distribution)
CREATE INDEX idx_decisions_outcome_metrics ON decisions (tenant_id, decided_at, outcome) INCLUDE (score);

CREATE TABLE feature_snapshots (
    transaction_id UUID NOT NULL,
    tenant_id UUID NOT NULL, 
    schema_hash VARCHAR(64) NOT NULL,
    features_json JSONB NOT NULL,
    feature_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (transaction_id, tenant_id),
    FOREIGN KEY (transaction_id, tenant_id) REFERENCES transactions(transaction_id, tenant_id)
) PARTITION BY LIST (tenant_id);

-- ==========================================
-- 6. Review & Mobile Feedback (Partitioned)
-- ==========================================
CREATE TABLE review_tasks (
    review_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'assigned', 'resolved')),
    assignee_id UUID,    
    severity VARCHAR(10) DEFAULT 'medium',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (review_id, tenant_id),
    FOREIGN KEY (transaction_id, tenant_id) REFERENCES transactions(transaction_id, tenant_id)
) PARTITION BY LIST (tenant_id);

CREATE TRIGGER update_reviews_modtime BEFORE UPDATE ON review_tasks FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Partial Index for extremely fast "Open Tasks" dashboard load
CREATE INDEX idx_reviews_pending_priority ON review_tasks (tenant_id, created_at DESC) WHERE status = 'pending';

-- Mobile Activity Log (Reviewer Performance)
CREATE TABLE review_activity_logs (
    log_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL, -- Partition Key
    review_id UUID NOT NULL,
    user_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL, -- "assign", "resolve", "comment"
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (log_id, tenant_id)
) PARTITION BY LIST (tenant_id);

CREATE TABLE labels (
    label_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    source VARCHAR(20) NOT NULL CHECK (source IN ('manual', 'chargeback', 'auto')),
    value VARCHAR(20) NOT NULL CHECK (value IN ('fraud', 'legit')),
    comment TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (label_id, tenant_id),
    FOREIGN KEY (transaction_id, tenant_id) REFERENCES transactions(transaction_id, tenant_id)
) PARTITION BY LIST (tenant_id);

-- ==========================================
-- 7. Analytics Aggregation (Trigger-based)
-- ==========================================
CREATE TABLE daily_metrics (
    tenant_id UUID NOT NULL,
    metric_date DATE NOT NULL,
    total_transactions BIGINT DEFAULT 0,
    total_approved BIGINT DEFAULT 0,
    total_blocked BIGINT DEFAULT 0,
    total_challenged BIGINT DEFAULT 0,
    total_fraud_value DECIMAL(18, 2) DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (tenant_id, metric_date)
);

-- Trigger function to update metrics continuously
CREATE OR REPLACE FUNCTION aggregate_decision_metrics()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO daily_metrics (tenant_id, metric_date, total_transactions, total_approved, total_blocked, total_challenged, updated_at)
    VALUES (
        NEW.tenant_id, 
        DATE(NEW.decided_at), 
        1, 
        CASE WHEN NEW.outcome = 'approve' THEN 1 ELSE 0 END,
        CASE WHEN NEW.outcome = 'block' THEN 1 ELSE 0 END,
        CASE WHEN NEW.outcome = 'challenge' THEN 1 ELSE 0 END,
        NOW()
    )
    ON CONFLICT (tenant_id, metric_date) DO UPDATE SET
        total_transactions = daily_metrics.total_transactions + 1,
        total_approved = daily_metrics.total_approved + EXCLUDED.total_approved,
        total_blocked = daily_metrics.total_blocked + EXCLUDED.total_blocked,
        total_challenged = daily_metrics.total_challenged + EXCLUDED.total_challenged,
        updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_daily_metrics
AFTER INSERT ON decisions
FOR EACH ROW
EXECUTE FUNCTION aggregate_decision_metrics();

-- ==========================================
-- 8. Audit Logging (Immutable System-wide)
-- ==========================================
CREATE TABLE audit_logs (
    event_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    tenant_id UUID, -- Nullable for system-wide events
    actor_id UUID,  -- User ID or System
    action VARCHAR(50) NOT NULL, -- e.g. "policy.update"
    entity_type VARCHAR(50),
    entity_id UUID,
    changes JSONB, -- Previous vs New values
    ip_address INET,
    occurred_at TIMESTAMPTZ NOT NULL,
    
    PRIMARY KEY (event_id, occurred_at)
) PARTITION BY RANGE (occurred_at);

-- Partition audit logs by month (Retention Strategy)
CREATE TABLE audit_logs_default PARTITION OF audit_logs DEFAULT;
