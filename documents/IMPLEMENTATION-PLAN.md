# Sentinal Core - Execution Kanban (v7.1)

**Total Tasks**: 80
**Methodology**: Atomic Kanban (Task <= 1 day)

---

## 🗄️ EPIC 2: Database & Migration (SC-011 → SC-022)
- [x] **SC-011**: Integrate `pgxpool` ✅
- [x] **SC-012**: Add DB timeout & config ✅
- [x] **SC-013**: Integrate `golang-migrate` ✅
- [x] **SC-014**: Test forward migration ✅
- [x] **SC-015**: Test rollback migration ✅
- [x] **SC-016**: Add `testcontainer` integration test ✅
- [x] **SC-017**: Implement Unit of Work pattern ✅
- [x] **SC-018**: Add DB transaction middleware ✅
- [x] **SC-019**: Implement Base Repository (Generics) ✅
- [x] **SC-020**: Implement Soft Delete logic ✅
- [x] **SC-021**: Handle DB errors and mapping (pkg/errors) – `0.5d` ✅ 2026-02-15
- [x] **SC-022**: Configure connection pool for high load – `0.5d` ✅ 2026-02-15

---

## 🔐 EPIC 3: Auth & RBAC (SC-023 → SC-032)
- [x] **SC-023**: Implement password hashing (`bcrypt`) ✅ (Formerly SC-021)
- [x] **SC-024**: Implement login endpoint ✅ (Formerly SC-022)
- [x] **SC-025**: Implement JWT issuance ✅ (Formerly SC-023)
- [x] **SC-026**: Implement JWT Auth middleware – `0.5d` ✅ 2026-02-15
- [ ] **SC-027**: Implement role permission matrix – `1d`
- ...

---

## 📜 EPIC 4: Transaction API (SC-033 → SC-042)
- [x] **SC-033**: Implement `POST /transactions` ✅ (Formerly SC-031)
- [x] **SC-034**: Implement Idempotency ✅ (Formerly SC-032)
- [ ] **SC-035**: Implement `GET /transactions` (Pagination) – `1d`
- ...
