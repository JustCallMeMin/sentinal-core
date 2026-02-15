# IMPLEMENTATION PLAN: SC-027 - Role-Permission Matrix

## 🎯 GOAL
Implement a robust Role-Based Access Control (RBAC) system that enforces permissions at the API level using JWT claims.

---

## 📋 REQUIREMENTS
| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| R1 | Seed core permissions (read/write/manage) | High | ⏳ |
| R2 | Seed default roles (Admin, Analyst, Viewer) | High | ⏳ |
| R3 | Fetch permissions associated with a user's role | High | ⏳ |
| R4 | Include permissions in JWT claims | High | ⏳ |
| R5 | Create middleware to enforce permission requirements | High | ⏳ |

---

## 🏗️ ARCHITECTURE & DATA MODEL
- **Permissions**: Slug-based (e.g., `transactions:read`).
- **Roles**: Linked to a tenant (or system-wide if `tenant_id` is null).
- **JWT**: Stateless verification via `permissions` claim.

---

## 🛠️ IMPLEMENTATION STEPS

### Step 1: Migration (Seeding)
- File: `migrations/000006_seed_roles_permissions.up.sql`
- Actions: 
    - Insert permissions: `transactions:read`, `transactions:write`, `users:manage`.
    - Insert system roles: `SUPER_ADMIN`, `ANALYST`, `VIEWER`.
    - Link permissions to roles in `role_permissions`.

### Step 2: Repository Layer
- File: `internal/repository/rbac_repository.go`
- Method: `GetPermissionsByRoleID(ctx, roleID int) ([]string, error)`
- Update `internal/repository/uow.go` to include `RBAC()`.

### Step 3: Auth Layer Updates
- File: `internal/api/auth/token_service.go`
    - Update `UserClaims` struct with `Permissions []string`.
    - Update `GenerateToken` signature and implementation.
- File: `internal/api/auth/service.go`
    - Fetch permissions in `Login` method before token generation.

### Step 4: Middleware Layer
- File: `internal/server/middleware_auth.go` (or new `middleware_rbac.go`)
    - Implement `RBACMiddleware(permissionSlug string) fiber.Handler`.
    - Extract permissions from `c.Locals`.

### Step 5: Integration & Verification
- Test: Update `tests/integration/api/auth_flow_test.go` to verify permissions in token.
- Test: Create a protected route and verify 403 Forbidden scenarios.

---

## ⚠️ RISKS & CONSTRAINTS
- **JWT Bloat**: We will only store permission slugs to minimize token size.
- **Null RoleID**: Handle users without assigned roles (default to lowest privilege or guest).

---

## ✅ ACCEPTANCE CRITERIA
- [ ] Users receive a JWT containing their permissions upon login.
- [ ] Middleware correctly allows access for users with valid permissions.
- [ ] Middleware returns `401 Unauthorized` for invalid tokens (handled by Auth middleware).
- [ ] Middleware returns `403 Forbidden` for users lacking required permissions.
