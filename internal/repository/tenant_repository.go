package repository

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinal/core/internal/domain/models"
)

type tenantRepository struct {
	*BaseRepository[models.Tenant]
}

func NewTenantRepository(pool *pgxpool.Pool) *tenantRepository {
	return &tenantRepository{
		BaseRepository: NewBaseRepository[models.Tenant](pool, "tenants"),
	}
}

func (r *tenantRepository) Create(ctx context.Context, t *models.Tenant) error {
	if t.TenantID == uuid.Nil {
		t.TenantID = uuid.New()
	}

	// 1. Insert into tenants table
	query := `
		INSERT INTO tenants (tenant_id, name, industry_segment, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	_, err := r.Pool.Exec(ctx, query, t.TenantID, t.Name, t.IndustrySegment, t.Settings)
	if err != nil {
		return fmt.Errorf("failed to insert tenant: %w", err)
	}

	// 2. SC-HARDENING: Call automation function to create partitions
	partitionQuery := `SELECT create_tenant_partition($1)`
	if _, err := r.Pool.Exec(ctx, partitionQuery, t.TenantID); err != nil {
		// Log error but maybe don't fail the whole creation?
		// Actually, for SaaS, the partition IS mandatory.
		return fmt.Errorf("failed to create tenant partitions: %w", err)
	}

	return nil
}

func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return r.BaseRepository.GetByID(ctx, id, "tenant_id")
}

func (r *tenantRepository) GetByAPIKeyHash(ctx context.Context, hash string) (*models.Tenant, error) {
	var t models.Tenant
	query := `SELECT * FROM tenants WHERE api_key_hash = $1 LIMIT 1`

	if err := pgxscan.Get(ctx, r.Pool, &t, query, hash); err != nil {
		return nil, fmt.Errorf("failed to get tenant by api key: %w", err)
	}

	return &t, nil
}

func (r *tenantRepository) List(ctx context.Context, limit, offset int) ([]*models.Tenant, error) {
	return r.BaseRepository.ListAll(ctx, limit, offset)
}

func (r *tenantRepository) Update(ctx context.Context, t *models.Tenant) error {
	query := `
		UPDATE tenants 
		SET name = $1, industry_segment = $2, settings = $3, updated_at = NOW()
		WHERE tenant_id = $4
	`
	_, err := r.Pool.Exec(ctx, query, t.Name, t.IndustrySegment, t.Settings, t.TenantID)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	return nil
}
