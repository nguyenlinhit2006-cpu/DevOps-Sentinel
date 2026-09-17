# Hướng Dẫn Khởi Chạy DevOps Sentinel Bằng Docker Trên Windows

Tài liệu này cung cấp hướng dẫn từng bước chi tiết để cài đặt, đóng gói và khởi chạy toàn bộ nền tảng **DevOps Sentinel** trên hệ điều hành **Windows (10 / 11)** thông qua Docker mà **không cần cài đặt thủ công** PHP, Composer, Go hay PostgreSQL.

---

## 1. Yêu cầu Hệ thống (Prerequisites)

Trước khi bắt đầu, máy tính Windows của bạn cần có:

1. **Hệ điều hành**: Windows 10 (Build 19041 trở lên) hoặc Windows 11 (64-bit).
2. **Docker Desktop for Windows**:
   - Đã cài đặt [Docker Desktop](https://www.docker.com/products/docker-desktop/).
   - Đã bật tính năng **WSL 2 Backend** trong phần *Settings > General > Use the WSL 2 based engine*.
3. **Phần cứng khuyến nghị**:
   - Tối thiểu 4GB RAM khả dụng cho Docker.
   - Ổ cứng trống tối thiểu 5GB để lưu trữ image và database.

---

## 2. Cách 1: Khởi chạy Nhanh Nhất (Chỉ 1 Container Duy Nhất)

Hình ảnh Docker **All-in-One** đã tích hợp sẵn:
- **Cơ sở dữ liệu**: PostgreSQL 16 (tự động tạo database `devops_sentinel`).
- **Backend API**: Laravel 11 RESTful API (chạy tại cổng `8000`).
- **Frontend UI**: Go WebAssembly thuần (phục vụ tại cổng `3000`).
- **Dữ liệu mẫu**: Tự động chạy migration và nạp tài khoản kiểm thử sẵn sàng trải nghiệm.

### Bước 1: Mở PowerShell hoặc Windows Terminal
Nhấn phím `Windows + X` và chọn **Terminal** hoặc **PowerShell**.

### Bước 2: Tự build image trên máy (nếu chưa có)
Di chuyển đến thư mục dự án `DevOps-Sentinel`:

```powershell
cd C:\path\to\DevOps-Sentinel
docker build -t devops-sentinel:latest .
```

### Bước 3: Khởi chạy Container
Chạy lệnh sau để khởi động toàn bộ hệ thống:

```powershell
# Chạy container và map các cổng 3000 và 8000 ra máy Windows
docker run -d -p 3000:3000 -p 8000:8000 --name devops-sentinel devops-sentinel:latest
```

> [!TIP]
> **Lưu trữ dữ liệu bền vững (Persistent Data)**:
> Nếu bạn muốn lưu giữ dữ liệu dự án và lượt build không bị mất khi xóa container, hãy gắn thêm volume:
> ```powershell
> docker run -d -p 3000:3000 -p 8000:8000 -v devops_sentinel_data:/var/lib/postgresql/data --name devops-sentinel devops-sentinel:latest
> ```

### Bước 4: Kiểm tra trạng thái hoạt động
Xem nhật ký khởi động của container:

```powershell
docker logs -f devops-sentinel
```

Khi thấy thông báo:
```text
🎉 [DevOps Sentinel] Toàn bộ hệ thống đã khởi chạy thành công!
   • Giao diện WebAssembly (Frontend): http://localhost:3000
   • Máy chủ RESTful API (Backend):   http://localhost:8000
```
Bạn có thể nhấn `Ctrl + C` để thoát màn hình xem log.

---

## 3. Trải Nghiệm Ứng Dụng Trên Trình Duyệt

Mở trình duyệt (Chrome, Edge, Firefox, Brave) trên Windows và truy cập:

🔗 **`http://localhost:3000`**

### Tài khoản Demo có sẵn (Đăng nhập 1-chạm):
Tại màn hình đăng nhập `http://localhost:3000/#/login`, hệ thống có sẵn thanh công cụ **"Chọn tài khoản thử nghiệm nhanh (1-chạm)"**:

| Vai trò | Email | Mật khẩu | Quyền hạn |
| :--- | :--- | :--- | :--- |
| 👑 **Admin** | `admin@devops-sentinel.local` | `Sentinel@123456` | Toàn quyền quản trị, quản lý nhóm, xóa dự án, xem Audit Logs. |
| 🛡️ **Team Lead** | `lead@devops-sentinel.local` | `Sentinel@123456` | Quản lý dự án, sinh lại secret, phân bổ thành viên, đóng cảnh báo. |
| 👁️ **Viewer** | `viewer@devops-sentinel.local` | `Sentinel@123456` | Xem Dashboard, danh sách dự án, lịch sử build và console logs. |

---

## 4. Các Lệnh Quản Lý Thường Dùng Trên Windows

### Tạm dừng container:
```powershell
docker stop devops-sentinel
```

### Khởi động lại container đã dừng:
```powershell
docker start devops-sentinel
```

### Xem log thời gian thực:
```powershell
docker logs -f devops-sentinel
```

### Xóa container để làm mới hoàn toàn:
```powershell
docker stop devops-sentinel
docker rm devops-sentinel
```

### Truy cập dòng lệnh bên trong container (Bash):
```powershell
docker exec -it devops-sentinel bash
```

---

## 5. Xử Lý Các Vấn Đề Thường Gặp (Troubleshooting trên Windows)

### 1. Lỗi xung đột cổng `3000` hoặc `8000`
- **Nguyên nhân**: Cổng 3000 hoặc 8000 đang được một ứng dụng khác (Node.js, Grafana...) sử dụng trên máy host Windows.
- **Khắc phục**: Đổi cổng host bên ngoài khi chạy Docker:
  ```powershell
  # Map cổng 3001 cho Frontend và 8001 cho Backend
  docker run -d -p 3001:3000 -p 8001:8000 --name devops-sentinel devops-sentinel:latest
  ```

### 2. Lỗi ký tự xuống dòng Windows (CRLF vs LF)
- **Nguyên nhân**: Khi clone code bằng Git trên Windows, cấu hình `core.autocrlf true` có thể tự động đổi ký tự xuống dòng của file shell script `docker-entrypoint.sh` từ `LF` sang `CRLF`, khiến Linux báo lỗi `exec format error` hoặc `/usr/bin/env: 'bash\r': no such file or directory`.
- **Khắc phục**: Chạy lệnh chuyển đổi sang `LF` trong PowerShell trước khi build:
  ```powershell
  (Get-Content docker-entrypoint.sh -Raw) -replace "`r`n", "`n" | Set-Content -NoNewline docker-entrypoint.sh
  ```
  Hoặc cấu hình Git:
  ```powershell
  git config --global core.autocrlf input
  ```

### 3. Docker Desktop ngốn nhiều RAM (WSL 2 Vmmem)
- **Khắc phục**: Tạo hoặc chỉnh sửa file cấu hình `%UserProfile%\.wslconfig` trên Windows:
  ```ini
  [wsl2]
  memory=4GB
  processors=2
  swap=2GB
  ```
  Sau đó mở PowerShell chạy:
  ```powershell
  wsl --shutdown
  ```
  và mở lại Docker Desktop.
