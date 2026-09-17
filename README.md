# DevOps Sentinel - Hệ Thống Giám Sát Triển Khai Tự Động (CI/CD Pipeline Monitoring Platform)

Hệ thống giám sát và quản lý trạng thái CI/CD Pipeline đa dự án tập trung, cảnh báo sự cố liên tiếp thời gian thực.

---

## 1. Công nghệ & Kiến trúc

- **Backend**: PHP 8.3 (Laravel 11 Pure API mode, kiến trúc stateless RESTful JSON).
- **Database**: PostgreSQL 16 (Hỗ trợ JSONB, chỉ mục tối ưu, quan hệ khóa ngoại toàn vẹn).
- **Frontend**: Go biên dịch sang WebAssembly (`GOOS=js GOARCH=wasm`), thao tác DOM trực tiếp qua thư viện bao bọc `syscall/js` độc lập.
- **Môi trường & Đóng gói**: Nix Flake (`flake.nix`, `flake.lock`) và Docker Compose (`docker-compose.yml`).

---

## 2. Hướng dẫn Khởi chạy Môi trường Cục bộ (Local Setup)

### Cách 1: Sử dụng Nix Flake (Khuyến nghị)
Môi trường đã được khóa chính xác phiên bản của PHP 8.3, Composer, Go, PostgreSQL CLI và Make trong `flake.nix` & `flake.lock`:

```bash
# Kích hoạt môi trường phát triển
nix develop
```

### Cách 2: Sử dụng các công cụ có sẵn trên máy
Yêu cầu hệ thống:
- PHP >= 8.3 (kèm extension: `pdo_pgsql`, `pgsql`, `mbstring`, `curl`, `openssl`, `zip`, `bcmath`, `pcntl`)
- Composer >= 2.7
- Docker & Docker Compose
- Go >= 1.22

---

## 3. Các bước Khởi chạy Backend

### Bước 1: Khởi động Cơ sở dữ liệu PostgreSQL
Sử dụng Docker Compose để khởi chạy container PostgreSQL 16 (được map vào cổng `5433:5432` trên máy host để không xung đột với các database khác):

```bash
make env-up
# hoặc: docker compose up -d postgres mailpit
```

### Bước 2: Cấu hình biến môi trường
File cấu hình `backend/.env` đã được thiết lập sẵn thông số kết nối:

```ini
DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5433
DB_DATABASE=devops_sentinel
DB_USERNAME=sentinel_user
DB_PASSWORD=sentinel_password

JWT_SECRET=devops_sentinel_super_secret_jwt_key_32bytes_min_2026
JWT_TTL=900
JWT_REFRESH_TTL=604800
```

### Bước 3: Chạy Migration & Nạp dữ liệu mẫu (Seed)
```bash
# Chạy migration tạo 7 bảng dữ liệu
make migrate

# Nạp dữ liệu mẫu kiểm thử
make seed
```

### Bước 4: Khởi chạy Máy chủ API Backend
```bash
make dev-backend
# API sẽ lắng nghe tại: http://127.0.0.1:8000
```

---

## 4. Tài khoản Mẫu Kiểm thử (Seeded Users)

Hệ thống đã nạp sẵn 3 tài khoản tương ứng với 3 cấp bậc phân quyền (RBAC):

| Vai trò | Email | Mật khẩu | Quyền hạn |
| :--- | :--- | :--- | :--- |
| **Admin** | `admin@devops-sentinel.local` | `Sentinel@123456` | Toàn quyền hệ thống, quản trị nhóm, xóa dự án, xem Audit Logs. |
| **Team Lead** | `lead@devops-sentinel.local` | `Sentinel@123456` | Tạo/sửa dự án, sinh lại webhook secret, quản lý thành viên nhóm, đóng cảnh báo. |
| **Viewer** | `viewer@devops-sentinel.local` | `Sentinel@123456` | Xem Dashboard, danh sách dự án, lịch sử pipeline và danh sách cảnh báo. |

---

## 5. Hướng dẫn Chạy Kiểm thử Tự động (Testing)

Hệ thống sở hữu bộ kiểm thử toàn diện cả Unit Test và Feature Test:

```bash
make test
# hoặc: cd backend && php artisan test
```

**Kết quả kiểm thử hiện tại:**
- **22/22 tests passed** (100%).
- **102 assertions**.

---

## 6. Danh sách Đối Chiếu Endpoints với API Contract Phase 2

Toàn bộ **27 endpoints** của hệ thống Backend đã hoàn thành 100%, đối chiếu chi tiết với Hợp đồng giao tiếp đã duyệt tại Phase 2:

| # | Phương thức | Đường dẫn API (Route) | Controller & Action | Bảo vệ / Quyền hạn | Đối chiếu API Contract Phase 2 |
| :---: | :---: | :--- | :--- | :--- | :---: |
| 1 | `GET` | `/api/v1/health` | Tuyến kiểm tra sức khỏe | Public | ✅ Khớp |
| 2 | `POST` | `/api/v1/auth/register` | `XacThucController@dangKy` | Public | ✅ Khớp |
| 3 | `POST` | `/api/v1/auth/login` | `XacThucController@dangNhap` | Public | ✅ Khớp |
| 4 | `POST` | `/api/v1/auth/refresh` | `XacThucController@lamMoiToken` | Public | ✅ Khớp |
| 5 | `GET` | `/api/v1/auth/me` | `XacThucController@thongTinCaNhan` | JWT Bearer | ✅ Khớp |
| 6 | `POST` | `/api/v1/auth/logout` | `XacThucController@dangXuat` | JWT Bearer | ✅ Khớp |
| 7 | `GET` | `/api/v1/admin/kiem-tra-quyen` | Tuyến test quyền Admin | JWT + `role:admin` | ✅ Khớp |
| 8 | `POST` | `/api/v1/webhooks/projects/{id}` | `WebhookController@tiepNhan` | Middleware HMAC SHA256 | ✅ Khớp |
| 9 | `GET` | `/api/v1/projects` | `DuAnController@danhSach` | JWT Bearer | ✅ Khớp |
| 10 | `POST` | `/api/v1/projects` | `DuAnController@taoMoi` | JWT + `role:team_lead` | ✅ Khớp |
| 11 | `GET` | `/api/v1/projects/{id}` | `DuAnController@chiTiet` | JWT Bearer | ✅ Khớp |
| 12 | `PUT` | `/api/v1/projects/{id}` | `DuAnController@capNhat` | JWT + `role:team_lead` | ✅ Khớp |
| 13 | `DELETE`| `/api/v1/projects/{id}` | `DuAnController@xoa` | JWT + `role:admin` | ✅ Khớp |
| 14 | `POST` | `/api/v1/projects/{id}/regenerate-secret` | `DuAnController@sinhLaiSecret` | JWT + `role:team_lead` | ✅ Khớp |
| 15 | `GET` | `/api/v1/projects/{projectId}/runs` | `LuotChayController@danhSachTheoDuAn` | JWT Bearer | ✅ Khớp |
| 16 | `GET` | `/api/v1/runs/{id}` | `LuotChayController@chiTiet` | JWT Bearer | ✅ Khớp |
| 17 | `GET` | `/api/v1/teams` | `NhomController@danhSach` | JWT Bearer | ✅ Khớp |
| 18 | `POST` | `/api/v1/teams` | `NhomController@taoMoi` | JWT + `role:admin` | ✅ Khớp |
| 19 | `GET` | `/api/v1/teams/{id}` | `NhomController@chiTiet` | JWT Bearer | ✅ Khớp |
| 20 | `PUT` | `/api/v1/teams/{id}` | `NhomController@capNhat` | JWT + `role:team_lead` | ✅ Khớp |
| 21 | `DELETE`| `/api/v1/teams/{id}` | `NhomController@xoa` | JWT + `role:admin` | ✅ Khớp |
| 22 | `POST` | `/api/v1/teams/{id}/members` | `NhomController@themThanhVien` | JWT + `role:team_lead` | ✅ Khớp |
| 23 | `DELETE`| `/api/v1/teams/{id}/members/{userId}` | `NhomController@xoaThanhVien` | JWT + `role:team_lead` | ✅ Khớp |
| 24 | `GET` | `/api/v1/dashboard/summary` | `BangDieuKhienController@tongQuan` | JWT Bearer | ✅ Khớp |
| 25 | `GET` | `/api/v1/dashboard/metrics` | `BangDieuKhienController@soLieuThongKe` | JWT Bearer | ✅ Khớp |
| 26 | `GET` | `/api/v1/alerts` | `CanhBaoController@danhSach` | JWT Bearer | ✅ Khớp |
| 27 | `POST` | `/api/v1/alerts/{id}/resolve` | `CanhBaoController@dongCanhBao` | JWT + `role:team_lead` | ✅ Khớp |
| 28 | `GET` | `/api/v1/audit-logs` | `NhatKyController@danhSach` | JWT + `role:admin` | ✅ Khớp |

---

## 7. Hướng dẫn Khởi chạy Frontend Go WebAssembly

Frontend của DevOps Sentinel được xây dựng hoàn toàn bằng **Go WebAssembly thuần** (`syscall/js` DOM Wrapper độc lập), không sử dụng bất kỳ thư viện hay framework JavaScript bên ngoài nào.

### Bước 1: Biên dịch mã nguồn Go sang WebAssembly
```bash
make build-wasm
# Lệnh sẽ sinh file: frontend/dist/main.wasm
```

### Bước 2: Khởi chạy Máy chủ Frontend Tĩnh
```bash
make dev-frontend
# hoặc: cd frontend && go run ./cmd/server
```
Ứng dụng sẽ sẵn sàng phục vụ tại: **`http://localhost:3000`**

### Trải nghiệm các tính năng đã hoàn thành (Phase 7 & Phase 8):
1. **Đăng nhập nhanh 1-chạm**: Truy cập `http://localhost:3000/#/login`, bấm vào một trong các nút Demo (`Admin`, `Team Lead`, `Viewer`) để tự động điền thông tin và đăng nhập.
2. **Dashboard Tổng quan thời gian thực**:
   - 4 thẻ KPI động: Tổng số dự án, Lượt chạy hôm nay, Tỷ lệ thành công (%), Cảnh báo Sentinel active.
   - Lưới danh sách dự án với ô tìm kiếm tức thì, thẻ thông tin lượt build gần nhất và huy hiệu trạng thái.
   - Bảng hoạt động 8 lượt chạy gần đây nhất trên toàn hệ thống.
   - Thanh Navbar hiển thị Avatar, Tên và Vai trò người dùng kèm hiệu ứng radar animation.