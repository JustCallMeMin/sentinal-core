-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For fuzzy search

-- 0. UTILITY FUNCTIONS
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 1. Identity & Tenant Layer
CREATE TABLE IF NOT EXISTS tenants (
    tenant_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    industry_segment VARCHAR(50) NOT NULL,
    api_key_hash TEXT,
    settings JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TRIGGER update_tenants_modtime BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- 2. SaaS User & RBAC
CREATE TABLE IF NOT EXISTS permissions (
    permission_id SERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS roles (
    role_id SERIAL PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    UNIQUE(tenant_id, name)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_id INT REFERENCES permissions(permission_id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id INT REFERENCES roles(role_id),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'locked')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

CREATE TRIGGER update_users_modtime BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- 3. Master Partitioned Tables
CREATE TABLE IF NOT EXISTS transactions (
    transaction_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    correlation_id VARCHAR(128),
    amount DECIMAL(18, 2) NOT NULL,
    currency CHAR(3) DEFAULT 'USD',
    occurred_at TIMESTAMPTZ NOT NULL,
    payload JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (tenant_id, transaction_id)
) PARTITION BY LIST (tenant_id);

CREATE TABLE IF NOT EXISTS feature_snapshots (
    transaction_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    schema_hash TEXT NOT NULL,
    features_json JSONB NOT NULL,
    feature_hash TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (tenant_id, transaction_id)
) PARTITION BY LIST (tenant_id);

CREATE TABLE IF NOT EXISTS labels (
    label_id UUID NOT NULL,
    transaction_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    source VARCHAR(20) NOT NULL,
    value VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (tenant_id, label_id)
) PARTITION BY LIST (tenant_id);

-- 4. Indices
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at ON transactions (occurred_at);
CREATE INDEX IF NOT EXISTS idx_transactions_correlation ON transactions (tenant_id, correlation_id);
CREATE INDEX IF NOT EXISTS idx_transactions_brin ON transactions USING brin (occurred_at);

-- 5. Automation Function
CREATE OR REPLACE FUNCTION create_tenant_partition(t_uuid UUID) RETURNS void AS $$
BEGIN
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF transactions FOR VALUES IN (%L)',
        'transactions_' || replace(t_uuid::text, '-', '_'), t_uuid
    );
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF feature_snapshots FOR VALUES IN (%L)',
        'features_' || replace(t_uuid::text, '-', '_'), t_uuid
    );
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF labels FOR VALUES IN (%L)',
        'labels_' || replace(t_uuid::text, '-', '_'), t_uuid
    );
END;
$$ LANGUAGE plpgsql;
