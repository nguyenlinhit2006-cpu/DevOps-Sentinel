//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"devops-sentinel/frontend/internal/api"
	"devops-sentinel/frontend/internal/dom"
	"devops-sentinel/frontend/internal/routing"
	"devops-sentinel/frontend/internal/state"
)

/**
 * Điểm khởi nhập (Entrypoint) của ứng dụng Frontend Go WebAssembly.
 * Thiết lập DOM, State, API Client, đăng ký router và giữ cho tiến trình WASM luôn chạy.
 */
func main() {
	fmt.Println("🚀 [DevOps Sentinel Frontend] Go WebAssembly runtime đã khởi chạy thành công!")

	// Kiểm tra sức khỏe kết nối Backend
	go func() {
		thongDiep, err := api.KiemTraSucKhoe()
		if err == nil {
			fmt.Printf("✅ Kết nối Backend thành công: %s\n", thongDiep)
		} else {
			fmt.Printf("⚠️ Chưa thể kết nối Backend API: %v\n", err)
		}
	}()

	// Khởi tạo bộ định tuyến Hash Routing
	bdh := routing.LayBoDieuHuong()

	// Tuyến đăng nhập
	bdh.DangKy("#/login", func(duongDan string, thamSo map[string]string) {
		renderGiaoDienDangNhap()
	})

	// Tuyến đăng ký
	bdh.DangKy("#/register", func(duongDan string, thamSo map[string]string) {
		renderGiaoDienDangKy()
	})

	// Tuyến Dashboard tổng quan
	bdh.DangKy("#/dashboard", func(duongDan string, thamSo map[string]string) {
		renderGiaoDienDashboard()
	})

	// Tuyến xử lý 404
	bdh.DangKyKhongTimThay(func(duongDan string, thamSo map[string]string) {
		khungChinh := dom.LayPhanTuTheoId("app")
		khungChinh.XoaHetCon()

		theThongBao := dom.TaoPhanTu("div").ThemLopCss("thong-bao-404")
		theThongBao.DatHtml(`
			<div class="card-404">
				<h2>404 - Không tìm thấy trang</h2>
				<p>Đường dẫn bạn yêu cầu không tồn tại hoặc đã bị xóa.</p>
				<a href="#/dashboard" class="btn btn-primary">Về Trang Chủ</a>
			</div>
		`)
		khungChinh.ThemCon(theThongBao)
	})

	// Khởi chạy bộ định tuyến
	bdh.KhoiChay()

	// Giữ vòng lặp runtime WASM luôn hoạt động
	select {}
}

/**
 * Render giao diện mẫu ban đầu cho trang Đăng nhập (sẽ hoàn thiện toàn diện ở Phase 8).
 */
func renderGiaoDienDangNhap() {
	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	theContainer := dom.TaoPhanTu("div").ThemLopCss("auth-container")
	theContainer.DatHtml(`
		<div class="auth-card">
			<div class="brand-header">
				<div class="brand-logo">🛡️</div>
				<h1>DevOps Sentinel</h1>
				<p>Hệ thống giám sát triển khai tự động</p>
			</div>
			<div class="form-wrapper">
				<div class="form-group">
					<label>Địa chỉ Email</label>
					<input type="email" id="input-email" placeholder="admin@devops-sentinel.local" value="admin@devops-sentinel.local" />
				</div>
				<div class="form-group">
					<label>Mật khẩu</label>
					<input type="password" id="input-password" placeholder="••••••••" value="Sentinel@123456" />
				</div>
				<div id="auth-alert" class="alert-box" style="display: none;"></div>
				<button id="btn-dang-nhap" class="btn btn-primary btn-block">Đăng Nhập</button>
				<div class="auth-footer">
					Chưa có tài khoản? <a href="#/register">Đăng ký ngay</a>
				</div>
			</div>
		</div>
	`)
	khungChinh.ThemCon(theContainer)

	// Gắn sự kiện nút đăng nhập
	dom.LayPhanTuTheoId("btn-dang-nhap").GanSuKien("click", func(this js.Value, args []js.Value) any {
		email := dom.LayPhanTuTheoId("input-email").LayGiaTri()
		matKhau := dom.LayPhanTuTheoId("input-password").LayGiaTri()

		theThongBao := dom.LayPhanTuTheoId("auth-alert")

		go func() {
			err := api.DangNhap(email, matKhau)
			if err != nil {
				theThongBao.DatThuocTinh("style", "display: block;").ThemLopCss("alert-error")
				theThongBao.DatNoiDung(fmt.Sprintf("❌ Lỗi: %v", err))
			} else {
				theThongBao.DatThuocTinh("style", "display: block;").ThemLopCss("alert-success")
				theThongBao.DatNoiDung("✅ Đăng nhập thành công! Đang chuyển hướng...")
				routing.ChuyenHuong("#/dashboard")
			}
		}()
		return nil
	})
}

/**
 * Render giao diện mẫu ban đầu cho trang Đăng ký (sẽ hoàn thiện ở Phase 8).
 */
func renderGiaoDienDangKy() {
	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	theContainer := dom.TaoPhanTu("div").ThemLopCss("auth-container")
	theContainer.DatHtml(`
		<div class="auth-card">
			<div class="brand-header">
				<div class="brand-logo">🛡️</div>
				<h1>Tạo Tài Khoản</h1>
				<p>DevOps Sentinel Monitoring</p>
			</div>
			<div class="form-wrapper">
				<div class="form-group">
					<label>Họ và Tên</label>
					<input type="text" id="input-name" placeholder="Nguyễn Văn A" />
				</div>
				<div class="form-group">
					<label>Email</label>
					<input type="email" id="input-email" placeholder="dev@company.local" />
				</div>
				<div class="form-group">
					<label>Mật khẩu</label>
					<input type="password" id="input-password" placeholder="Tối thiểu 6 ký tự" />
				</div>
				<div class="form-group">
					<label>Xác nhận Mật khẩu</label>
					<input type="password" id="input-password-confirm" placeholder="Nhập lại mật khẩu" />
				</div>
				<div id="auth-alert" class="alert-box" style="display: none;"></div>
				<button id="btn-dang-ky" class="btn btn-primary btn-block">Đăng Ký</button>
				<div class="auth-footer">
					Đã có tài khoản? <a href="#/login">Đăng nhập</a>
				</div>
			</div>
		</div>
	`)
	khungChinh.ThemCon(theContainer)

	dom.LayPhanTuTheoId("btn-dang-ky").GanSuKien("click", func(this js.Value, args []js.Value) any {
		ten := dom.LayPhanTuTheoId("input-name").LayGiaTri()
		email := dom.LayPhanTuTheoId("input-email").LayGiaTri()
		matKhau := dom.LayPhanTuTheoId("input-password").LayGiaTri()
		xacNhan := dom.LayPhanTuTheoId("input-password-confirm").LayGiaTri()
		theThongBao := dom.LayPhanTuTheoId("auth-alert")

		go func() {
			err := api.DangKy(ten, email, matKhau, xacNhan)
			if err != nil {
				theThongBao.DatThuocTinh("style", "display: block;").ThemLopCss("alert-error")
				theThongBao.DatNoiDung(fmt.Sprintf("❌ Lỗi: %v", err))
			} else {
				theThongBao.DatThuocTinh("style", "display: block;").ThemLopCss("alert-success")
				theThongBao.DatNoiDung("✅ Đăng ký thành công! Đang chuyển hướng...")
				routing.ChuyenHuong("#/dashboard")
			}
		}()
		return nil
	})
}

/**
 * Render khung cơ sở cho màn hình Dashboard (sẽ nạp dữ liệu thật ở Phase 8).
 */
func renderGiaoDienDashboard() {
	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	tt := state.LayTrangThai()
	tenNguoiDung := "Người dùng"
	vaiTro := "viewer"
	if tt.NguoiDungHienTai != nil {
		tenNguoiDung = tt.NguoiDungHienTai.Name
		vaiTro = tt.NguoiDungHienTai.Role
	}

	theContainer := dom.TaoPhanTu("div").ThemLopCss("dashboard-layout")
	theContainer.DatHtml(fmt.Sprintf(`
		<header class="app-navbar">
			<div class="nav-brand">
				<span class="logo">🛡️</span>
				<span class="title">DevOps Sentinel</span>
				<span class="badge badge-wasm">WASM</span>
			</div>
			<div class="nav-user">
				<span class="user-greeting">Xin chào, <strong>%s</strong></span>
				<span class="badge badge-role">%s</span>
				<button id="btn-dang-xuat" class="btn btn-outline btn-sm">Đăng xuất</button>
			</div>
		</header>
		<main class="app-content">
			<div class="welcome-banner card">
				<h2>⚡ Nền tảng Go WebAssembly đã sẵn sàng (Phase 7 Hoàn tất)</h2>
				<p>Hệ thống đã khởi tạo thành công lớp WebAssembly DOM Engine, Bộ định tuyến Hash Routing, Lớp API Client với cơ chế Silent Refresh và State Manager.</p>
			</div>
		</main>
	`, tenNguoiDung, vaiTro))

	khungChinh.ThemCon(theContainer)

	dom.LayPhanTuTheoId("btn-dang-xuat").GanSuKien("click", func(this js.Value, args []js.Value) any {
		go func() {
			api.DangXuat()
			routing.ChuyenHuong("#/login")
		}()
		return nil
	})
}
