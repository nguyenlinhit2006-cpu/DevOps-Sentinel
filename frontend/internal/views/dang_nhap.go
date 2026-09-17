//go:build js && wasm

package views

import (
	"fmt"
	"strings"
	"syscall/js"

	"devops-sentinel/frontend/internal/api"
	"devops-sentinel/frontend/internal/dom"
	"devops-sentinel/frontend/internal/routing"
	"devops-sentinel/frontend/internal/state"
)

/**
 * Hiển thị màn hình Đăng nhập tài khoản hệ thống.
 * Bao gồm form xác thực, thanh chọn nhanh tài khoản mẫu và xử lý lỗi chi tiết.
 */
func HienThiTrangDangNhap() {
	// Kiểm tra nếu đã có phiên hợp lệ thì chuyển hướng ngay vào Dashboard
	if state.LayTrangThai().DaDangNhap() {
		routing.ChuyenHuong("#/dashboard")
		return
	}

	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	theContainer := dom.TaoPhanTu("div").ThemLopCss("auth-container")
	theContainer.DatHtml(`
		<div class="auth-card">
			<div class="brand-header">
				<div class="brand-logo-wrap">
					<div class="brand-logo">🛡️</div>
					<div class="brand-glow"></div>
				</div>
				<h1>DevOps Sentinel</h1>
				<p>Hệ thống giám sát triển khai tự động CI/CD</p>
			</div>

			<!-- Thanh chọn nhanh tài khoản Demo hỗ trợ kiểm thử -->
			<div class="demo-accounts-box">
				<span class="demo-title">⚡ Chọn nhanh tài khoản Demo:</span>
				<div class="demo-buttons">
					<button type="button" class="btn-demo" id="demo-admin" title="Quyền Quản trị viên cao nhất">
						<span class="demo-role admin">Admin</span>
					</button>
					<button type="button" class="btn-demo" id="demo-lead" title="Quyền Trưởng nhóm kỹ thuật">
						<span class="demo-role lead">Team Lead</span>
					</button>
					<button type="button" class="btn-demo" id="demo-viewer" title="Quyền Chỉ xem">
						<span class="demo-role viewer">Viewer</span>
					</button>
				</div>
			</div>

			<form id="form-dang-nhap" class="form-wrapper" onsubmit="return false;">
				<div class="form-group">
					<label for="input-email">Địa chỉ Email</label>
					<div class="input-with-icon">
						<span class="input-icon">✉️</span>
						<input type="email" id="input-email" placeholder="tenban@congty.com" value="admin@devops-sentinel.local" required autocomplete="email" />
					</div>
				</div>

				<div class="form-group">
					<label for="input-password">Mật khẩu</label>
					<div class="input-with-icon">
						<span class="input-icon">🔒</span>
						<input type="password" id="input-password" placeholder="••••••••" value="Sentinel@123456" required autocomplete="current-password" />
					</div>
				</div>

				<div id="auth-alert" class="alert-box" style="display: none;"></div>

				<button type="submit" id="btn-dang-nhap" class="btn btn-primary btn-block">
					<span class="btn-text">Đăng Nhập</span>
				</button>

				<div class="auth-footer">
					Chưa có tài khoản? <a href="#/register">Đăng ký thành viên mới</a>
				</div>
			</form>
		</div>
	`)
	khungChinh.ThemCon(theContainer)

	// Xử lý sự kiện bấm nút Demo Accounts
	ganSuKienTaiKhoanDemo := func(idNut, email, matKhau string) {
		nut := dom.LayPhanTuTheoId(idNut)
		if nut.HopLe() {
			nut.GanSuKien("click", func(this js.Value, args []js.Value) any {
				dom.LayPhanTuTheoId("input-email").DatGiaTri(email)
				dom.LayPhanTuTheoId("input-password").DatGiaTri(matKhau)
				hienThiThongBaoAuth("info", fmt.Sprintf("ℹ️ Đã nạp tài khoản: %s", email))
				return nil
			})
		}
	}

	ganSuKienTaiKhoanDemo("demo-admin", "admin@devops-sentinel.local", "Sentinel@123456")
	ganSuKienTaiKhoanDemo("demo-lead", "lead@devops-sentinel.local", "Sentinel@123456")
	ganSuKienTaiKhoanDemo("demo-viewer", "viewer@devops-sentinel.local", "Sentinel@123456")

	// Xử lý đăng nhập
	xuLyDangNhap := func() {
		email := strings.TrimSpace(dom.LayPhanTuTheoId("input-email").LayGiaTri())
		matKhau := dom.LayPhanTuTheoId("input-password").LayGiaTri()

		if email == "" || matKhau == "" {
			hienThiThongBaoAuth("error", "Vui lòng nhập đầy đủ Email và Mật khẩu.")
			return
		}

		nutDangNhap := dom.LayPhanTuTheoId("btn-dang-nhap")
		nutDangNhap.DatThuocTinh("disabled", "true")
		nutDangNhap.DatHtml(`<span class="spinner-sm"></span> Đang xác thực...`)

		go func() {
			err := api.DangNhap(email, matKhau)
			if err != nil {
				nutDangNhap.XoaThuocTinh("disabled")
				nutDangNhap.DatHtml(`<span class="btn-text">Đăng Nhập</span>`)
				hienThiThongBaoAuth("error", fmt.Sprintf("❌ Đăng nhập thất bại: %v", err))
			} else {
				hienThiThongBaoAuth("success", "✅ Xác thực thành công! Đang chuyển hướng...")
				routing.ChuyenHuong("#/dashboard")
			}
		}()
	}

	dom.LayPhanTuTheoId("btn-dang-nhap").GanSuKien("click", func(this js.Value, args []js.Value) any {
		xuLyDangNhap()
		return nil
	})
}

/**
 * Hiển thị thông báo trạng thái trong khung Auth.
 */
func hienThiThongBaoAuth(loai string, noiDung string) {
	theThongBao := dom.LayPhanTuTheoId("auth-alert")
	if !theThongBao.HopLe() {
		return
	}

	theThongBao.XoaLopCss("alert-error")
	theThongBao.XoaLopCss("alert-success")
	theThongBao.XoaLopCss("alert-info")

	switch loai {
	case "error":
		theThongBao.ThemLopCss("alert-error")
	case "success":
		theThongBao.ThemLopCss("alert-success")
	default:
		theThongBao.ThemLopCss("alert-info")
	}

	theThongBao.DatThuocTinh("style", "display: block;")
	theThongBao.DatNoiDung(noiDung)
}
