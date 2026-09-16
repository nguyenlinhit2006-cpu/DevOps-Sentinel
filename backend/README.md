# DevOps Sentinel - Backend API Service

Dịch vụ RESTful API lõi của hệ thống giám sát DevOps Sentinel, xây dựng trên nền tảng PHP 8.3 và Laravel 11.

---

## 1. Yêu cầu & Cấu hình môi trường

- **PHP**: >= 8.3
- **Database**: PostgreSQL >= 16 (Cổng `5433` trong Docker Compose)
- **Tập tin cấu hình**: `.env` (Đã cấu hình sẵn thông số kết nối PostgreSQL và biến bí mật JWT)

## 2. Các lệnh thao tác nhanh

```bash
# Chạy migration cơ sở dữ liệu
php artisan migrate

# Nạp dữ liệu mẫu kiểm thử
php artisan db:seed

# Khởi chạy máy chủ phát triển
php artisan serve --port=8000

# Chạy kiểm thử tự động (Unit + Feature tests)
php artisan test
```

## 3. Danh sách kiểm thử tự động
- `tests/Unit/DichVuJwtTest.php`: Kiểm thử mã hóa / giải mã token JWT, đối soát chữ ký HMAC-SHA256.
- `tests/Unit/DichVuCanhBaoTest.php`: Kiểm thử đếm số lần lỗi liên tiếp của pipeline.
- `tests/Feature/XacThucVaPhanQuyenTest.php`: Kiểm thử Đăng ký, Đăng nhập, Profile, Refresh Token, Đăng xuất, Phân quyền RBAC.
- `tests/Feature/NghiepVuChinhTest.php`: Kiểm thử CRUD Dự án/Nhóm, Webhook Ingestion HMAC, Sentinel Alerting Rule, Dashboard Summary & Metrics, Audit Log.
