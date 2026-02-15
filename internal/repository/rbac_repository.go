package repository

import (
	"context"
	"fmt"

	"github.com/sentinal/core/internal/domain/repositories"
)

type rbacRepository struct {
	db Querier
}

func NewRBACRepository(db Querier) repositories.RBACRepository {
	return &rbacRepository{db: db}
}

func (r *rbacRepository) GetPermissionsByRoleID(ctx context.Context, roleID int) ([]string, error) {
	query := `
		SELECT p.slug
		FROM permissions p
		JOIN role_permissions rp ON p.permission_id = rp.permission_id
		WHERE rp.role_id = $1
	`

	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, fmt.Errorf("failed to scan permission slug: %w", err)
		}
		permissions = append(permissions, slug)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return permissions, nil
}
