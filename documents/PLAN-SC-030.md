# IMPLEMENTATION PLAN: SC-030 & Auth Hardening

Goal: Enhance security through configurable lockout, IP rate limiting, and MFA (TOTP) support.

## 1. Configurable Lockout (Upgrade)
- [x] **Step 1.1**: Update `pkg/config/config.go`:
    - Add `AuthMaxFailedAttempts int` (default 5).
    - Add `AuthLockoutMinutes int` (default 15).
- [x] **Step 1.2**: Update `internal/api/auth/service.go` to use `config.AuthLockoutMinutes`.
- [x] **Step 1.3**: Update `internal/repository/user_repository.go`:
    - Update `IncrementFailedAttempts` to use `maxFailedAttempts` from config (pass as arg).

## 2. IP-based Rate Limiting (Upgrade)
- [x] **Step 2.1**: Set up Redis client:
    - Update `internal/server/server.go` to include `Redis *redis.Client`.
    - Initialize Redis in `cmd/api/main.go` and pass to `server.New`.
- [x] **Step 2.2**: Implement Limiter Middleware in `internal/server/middleware_limit.go` using Fiber Limiter with Redis storage.
- [x] **Step 2.3**: Apply Limiter to `/api/v1/auth/login` in `routes.go`.

## 3. MFA Implementation (SC-030)
- [x] **Step 3.1**: Database Migration:
    - Create `migrations/000008_add_mfa_columns.up.sql` to add `mfa_enabled` and `mfa_secret` to `users`.
- [x] **Step 3.2**: Update `internal/domain/models/user.go`:
    - Add `MFAEnabled bool` and `MFASecret *string`.
- [x] **Step 3.3**: Update `AuthService.Login` flow:
    - If `MFAEnabled` and password correct -> Return `Status: "mfa_required"`, `MFAToken: "..."`.
    - Create temporary token containing UserID/TenantID for MFA phase.
- [x] **Step 3.4**: Implement MFA Setup endpoint:
    - `POST /api/v1/auth/mfa/setup`: Generate secret, return provisioning URI.
- [x] **Step 3.5**: Implement MFA Activate endpoint:
    - `POST /api/v1/auth/mfa/activate`: Validate 1st code, enable MFA in DB.
- [x] **Step 3.6**: Implement MFA Verify endpoint:
    - `POST /api/v1/auth/mfa/verify`: Final code verification during login.

## 4. Testing
- [x] **Step 4.1**: Unit tests for config loading.
- [x] **Step 4.2**: Integration tests for 5-attempt lockout (configurable).
- [x] **Step 4.3**: Integration tests for MFA login flow.

## Risks
- Redis connectivity down: Need graceful fallback for rate limiting (fallback to in-memory).
- MFA recovery: No recovery codes planned yet (marked as future enhancement).
