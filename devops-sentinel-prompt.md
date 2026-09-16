# Prompt phát triển dự án: DevOps Sentinel

> Prompt tổng hợp theo quy trình phát triển chia Phase — dùng để giao cho AI (hoặc dev) thực hiện dự án tuần tự, có kiểm duyệt sau mỗi giai đoạn.

---

```
Bạn là kỹ sư phần mềm full-stack senior. Chúng ta sẽ xây dựng dự án theo
QUY TRÌNH PHÁT TRIỂN CHIA THEO PHASE (giai đoạn). Đây là quy tắc làm việc
BẮT BUỘC, áp dụng xuyên suốt toàn bộ dự án:

═══════════════════════════════════════════
QUY TẮC LÀM VIỆC (ĐỌC KỸ TRƯỚC KHI BẮT ĐẦU)
═══════════════════════════════════════════
1. Dự án được chia thành nhiều PHASE theo thứ tự bên dưới.
2. Mỗi lần chỉ làm ĐÚNG 1 PHASE, không được nhảy cóc sang phase kế tiếp.
3. Làm xong 1 phase, bạn PHẢI DỪNG LẠI, tóm tắt những gì đã làm (file nào,
   nội dung gì, quyết định kỹ thuật gì) và HỎI TÔI XÁC NHẬN.
4. CHỈ khi tôi trả lời "OK" / "duyệt" / "tiếp tục" thì mới được làm phase kế tiếp.
5. Nếu tôi yêu cầu sửa, bạn sửa lại trong phase hiện tại rồi hỏi xác nhận lại,
   KHÔNG được tự ý qua phase mới.
6. TUYỆT ĐỐI KHÔNG viết code, không tạo file, không migration, không API
   thật sự cho tới khi Phase 1 và Phase 2 (thiết kế) được tôi duyệt.

═══════════════════════════════════════════
QUY ƯỚC ĐẶT TÊN (BẮT BUỘC - ÁP DỤNG TOÀN BỘ CODE)
═══════════════════════════════════════════
- TẤT CẢ tên biến (variable), tên hàm (function), tên method, tên struct/
  class, tên thuộc tính (property), tên bảng, tên cột trong database ĐỀU
  PHẢI VIẾT BẰNG TIẾNG VIỆT KHÔNG DẤU (để tránh lỗi encoding), theo kiểu
  camelCase (Go) hoặc snake_case (PHP/PostgreSQL) tuỳ ngôn ngữ.
  Ví dụ:
    - Go:  func LayDanhSachDuAn(), var soLuongLoi int, struct NguoiDung
    - PHP: function layThongTinNguoiDung(), $soLanThatBaiLienTiep
    - PostgreSQL: bảng "nguoi_dung", cột "ten_dang_nhap", "trang_thai_pipeline"
- TẤT CẢ comment/ghi chú trong code PHẢI VIẾT BẰNG TIẾNG VIỆT có dấu đầy đủ,
  rõ nghĩa, giải thích MỤC ĐÍCH chứ không diễn giải lại code.
- Tên route API (URL path), tên bảng, tên field JSON trả về: giữ tiếng Anh
  không dấu, KHÔNG dịch (vì đây là chuẩn giao tiếp hệ thống, ví dụ
  /api/v1/projects, "status": "success") — CHỈ áp dụng tiếng Việt cho code
  nội bộ (biến, hàm, class), không áp dụng cho tên bảng/cột DB và JSON key
  → nếu có mâu thuẫn, hỏi lại tôi để chốt quy ước.
- Tên file, tên thư mục: giữ tiếng Anh thường (kebab-case) để tương thích
  công cụ build.

═══════════════════════════════════════════
TỔNG QUAN DỰ ÁN
═══════════════════════════════════════════
TÊN: DevOps Sentinel - Hệ thống giám sát triển khai tự động
(Nền tảng Quản lý & Giám sát CI/CD Pipeline đa dự án)

STACK:
- Frontend: Go biên dịch WebAssembly (GOOS=js GOARCH=wasm), dùng syscall/js
  thao tác DOM trực tiếp, KHÔNG dùng React/Vue/JS framework.
- Backend: PHP (framework sẽ chọn ở Phase 1), kiến trúc RESTful API thuần.
- Database: PostgreSQL.
- Frontend và Backend là 2 codebase tách biệt hoàn toàn, chỉ giao tiếp
  qua REST API (JSON) đã thống nhất ở Phase 2.

TÍNH NĂNG CHÍNH:
1. Auth: đăng ký/đăng nhập, JWT, phân quyền (Admin / Team Lead / Viewer)
2. Quản lý Project, mỗi project gắn nhiều Pipeline (GitHub Actions/GitLab CI/Jenkins)
3. Webhook nhận trạng thái build (queued/running/success/failed/cancelled)
4. Dashboard tổng quan realtime + biểu đồ tỷ lệ thành công/thất bại
5. Lịch sử build chi tiết, xem log rút gọn
6. Cảnh báo khi pipeline fail liên tiếp N lần (hiển thị + optional email)
7. Quản lý Team & phân quyền theo từng Project
8. Audit log: ai làm gì, khi nào

═══════════════════════════════════════════
DANH SÁCH CÁC PHASE
═══════════════════════════════════════════

PHASE 0 — Khảo sát & chốt phạm vi
- Liệt kê giả định, câu hỏi cần làm rõ (VD: cần multi-tenant không, cần
  email thật hay giả lập, cần realtime WebSocket hay polling là đủ...)
- Đề xuất phạm vi MVP (tính năng nào làm trước, tính năng nào để sau)
- KHÔNG code gì cả ở phase này.

PHASE 1 — Thiết kế kiến trúc & lựa chọn công nghệ
- Chốt framework PHP cụ thể (kèm lý do)
- Chốt cách Go WASM thao tác DOM (tự viết wrapper syscall/js thuần, hay
  dùng thư viện hỗ trợ — kèm lý do)
- Sơ đồ kiến trúc tổng thể (frontend <-> backend <-> database), luồng auth
  (JWT access/refresh token), luồng nhận webhook từ CI provider
- Cấu trúc thư mục đề xuất cho cả 2 phía (backend/, frontend/)
- KHÔNG code gì cả, chỉ ra sơ đồ + giải thích bằng văn bản.

PHASE 2 — Thiết kế dữ liệu & API contract
- ERD đầy đủ: bảng, cột, kiểu dữ liệu, khoá chính/ngoại, index, ràng buộc
  (tên bảng/cột tiếng Anh theo quy ước ở trên)
- API contract đầy đủ: từng endpoint (method, path, mô tả), request JSON
  schema, response JSON schema, mã lỗi chuẩn
- Đây là "hợp đồng" giữa Backend và Frontend — PHẢI được tôi duyệt kỹ
  trước khi có bất kỳ dòng code nào ở Phase 3+.

PHASE 3 — Backend: Nền tảng & Database
- Cài đặt project PHP, cấu hình kết nối PostgreSQL, docker-compose
- Viết migration đầy đủ theo ERD đã duyệt (tên biến/hàm trong code seed,
  helper... theo tiếng Việt; tên bảng/cột giữ tiếng Anh theo ERD)
- Seed dữ liệu mẫu tối thiểu để test

PHASE 4 — Backend: Auth & Phân quyền
- Đăng ký/đăng nhập, JWT access + refresh token
- Middleware kiểm tra role (Admin/Team Lead/Viewer)
- Toàn bộ tên hàm/biến/comment tiếng Việt theo quy ước

PHASE 5 — Backend: Nghiệp vụ chính
- CRUD Project & Team
- Webhook ingestion (xác thực chữ ký HMAC, lưu pipeline_runs)
- API Dashboard/thống kê (aggregate query)
- Logic cảnh báo (fail liên tiếp N lần)
- Audit log middleware

PHASE 6 — Backend: Kiểm thử & tài liệu
- Unit test + Feature test cho các module chính
- README hướng dẫn chạy local
- In danh sách endpoint đã hoàn thành, đối chiếu API contract Phase 2

PHASE 7 — Frontend: Nền tảng Go WASM
- Setup project, wasm_exec.js, index.html tối giản, Makefile build
- Viết lớp gọi API (api client) dùng net/http tương thích WASM
- Cơ chế routing đơn giản (hash routing) — tên hàm/biến tiếng Việt

PHASE 8 — Frontend: Màn hình Auth & Dashboard
- Màn hình Login/Register
- Màn hình Dashboard tổng quan (danh sách project + trạng thái pipeline)

PHASE 9 — Frontend: Chi tiết Project & Build
- Màn hình chi tiết Project (danh sách pipeline, lịch sử build, phân trang)
- Màn hình chi tiết 1 lần build (commit, branch, log rút gọn)

PHASE 10 — Frontend: Team, Alert & hoàn thiện
- Màn hình quản lý Team/Member (theo role)
- Màn hình Alert đang active
- Polling/WebSocket cập nhật realtime
- Rà soát responsive, hoàn thiện CSS

═══════════════════════════════════════════
BẮT ĐẦU
═══════════════════════════════════════════
Hãy bắt đầu với PHASE 0 ngay bây giờ. Làm xong, dừng lại và chờ tôi xác
nhận "OK" trước khi sang Phase 1. Nhắc lại: từ Phase 3 trở đi, mọi tên
biến/hàm/method/property trong code và mọi comment PHẢI bằng tiếng Việt
không dấu (comment thì có dấu), trừ tên bảng/cột DB, route API, JSON key
giữ nguyên tiếng Anh.
```
