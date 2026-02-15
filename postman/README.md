# Sentinal Core - Postman Smoke Test

## 📦 Files

- `sentinal-core-smoke-test.postman_collection.json` - Collection chứa tất cả test cases
- `sentinal-core-local.postman_environment.json` - Environment cho local development

## 🚀 Setup Instructions

### 1. Import vào Postman

1. Mở Postman
2. Click **Import** (góc trên bên trái)
3. Drag & drop hoặc chọn cả 2 files:
   - `sentinal-core-smoke-test.postman_collection.json`
   - `sentinal-core-local.postman_environment.json`

### 2. Cấu hình Environment

1. Chọn environment **"Sentinal Core - Local"** (góc trên bên phải)
2. Click vào icon **eye** (👁️) để xem variables
3. Click **Edit** và cập nhật:
   - `tenant_id`: Thay `REPLACE_WITH_ACTUAL_TENANT_UUID` bằng UUID của tenant test

**Lấy Tenant ID**:
```bash
# Kết nối vào DB
docker exec -it <postgres_container> psql -U postgres -d sentinal_core

# Query tenant
SELECT tenant_id, name FROM tenants LIMIT 1;

# Hoặc tạo tenant mới
INSERT INTO tenants (name, industry_segment) 
VALUES ('Test Merchant', 'E-commerce') 
RETURNING tenant_id;
```

### 3. Chạy Server

```bash
# Từ thư mục gốc project
make run

# Hoặc
go run cmd/api/main.go
```

Server sẽ chạy tại `http://localhost:8080`

### 4. Chạy Tests

#### Chạy từng request:
1. Mở collection **"Sentinal Core - Smoke Test"**
2. Chọn request muốn test
3. Click **Send**
4. Xem kết quả ở tab **Test Results**

#### Chạy toàn bộ collection:
1. Click vào collection **"Sentinal Core - Smoke Test"**
2. Click **Run** (hoặc **Runner**)
3. Chọn environment **"Sentinal Core - Local"**
4. Click **Run Sentinal Core - Smoke Test**

---

## 📋 Test Cases

### Health Checks (3 tests)
- ✅ Root Endpoint - Verify service info
- ✅ Health Check - Verify liveness
- ✅ Readiness Check - Verify DB connection

### Transaction API (6 tests)
- ✅ **Create Transaction - Minimal** (Recommended) - Only required fields
- ✅ **Create Transaction - Full** - With optional fields (device_id, payload)
- ✅ Create Transaction - Validation Error (Negative amount)
- ✅ Create Transaction - Missing Tenant ID (401)
- ✅ Create Transaction - Invalid Currency (400)
- ✅ Create Transaction - Missing Required Field (400)

---

## 🎯 Request Examples

### ✨ Minimal Request (Recommended)
```json
POST /api/v1/transactions
Headers:
  Content-Type: application/json
  X-Tenant-ID: <your-tenant-uuid>

Body:
{
  "user_id": "user_12345",
  "amount": 150.50,
  "currency": "USD"
}
```

**Note**: 
- ✅ `ip_address` được server tự động extract từ request
- ✅ `device_id` là optional
- ✅ `payload` là optional

### 🔧 Full Request (With Optional Fields)
```json
{
  "user_id": "user_67890",
  "amount": 299.99,
  "currency": "USD",
  "device_id": "device_fingerprint_abc123",
  "payload": {
    "item_category": "electronics",
    "shipping_method": "express"
  }
}
```

---

## 🎯 Expected Results

**All tests should PASS** khi:
- Server đang chạy
- Database đã migrate
- Tenant ID hợp lệ trong environment

## 🔍 Troubleshooting

### Error: "Missing X-Tenant-ID header"
- Kiểm tra environment variable `tenant_id` đã được set chưa
- Verify tenant tồn tại trong database

### Error: "database connection lost"
- Kiểm tra PostgreSQL container đang chạy
- Verify connection string trong `.env`

### Error: "Validation failed"
- Kiểm tra request body format
- Verify required fields (user_id, amount, currency)

## 📊 Test Coverage

| Feature | Status |
|---------|--------|
| Health Checks | ✅ Complete |
| Transaction Ingestion | ✅ Complete (SC-031) |
| Transaction List | ⏳ Pending (SC-033) |
| Idempotency | ⏳ Pending (SC-032) |

---

**Last Updated**: 2026-02-15 (Epic 3 - SC-031)
