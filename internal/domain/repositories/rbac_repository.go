package repositories

import "context"

type RBACRepository interface {
	GetPermissionsByRoleID(ctx context.Context, roleID int) ([]string, error)
}
