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
 * Hiển thị màn hình Đăng ký tài khoản thành viên mới.
 */
func HienThiTrangDangKy() {
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
				<h1>Tạo Tài Khoản Mới</h1>
				<p>Gia nhập DevOps Sentinel để giám sát CI/CD toàn diện</p>
			</div>

			<form id="form-dang-ky" class="form-wrapper" onsubmit="return false;">
				<div class="form-group">
					<label for="input-name">Họ và Tên</label>
					<div class="input-with-icon">
						<span class="input-icon">👤</span>
						<input type="text" id="input-name" placeholder="Nguyễn Văn A" required />
					</div>
				</div>

				<div class="form-group">
					<label for="input-email">Địa chỉ Email</label>
					<div class="input-with-icon">
						<span class="input-icon">✉️</span>
						<input type="email" id="input-email" placeholder="dev@congty.com" required autocomplete="email" />
					</div>
				</div>

				<div class="form-group">
					<label for="input-password">Mật khẩu</label>
					<div class="input-with-icon">
						<span class="input-icon">🔒</span>
						<input type="password" id="input-password" placeholder="Tối thiểu 6 ký tự" required autocomplete="new-password" />
					</div>
				</div>

				<div class="form-group">
					<label for="input-password-confirm">Xác nhận Mật khẩu</label>
					<div class="input-with-icon">
						<span class="input-icon">🛡️</span>
						<input type="password" id="input-password-confirm" placeholder="Nhập lại mật khẩu trên" required autocomplete="new-password" />
					</div>
				</div>

				<div id="auth-alert" class="alert-box" style="display: none;"></div>

				<button type="submit" id="btn-dang-ky" class="btn btn-primary btn-block">
					<span class="btn-text">Tạo Tài Khoản</span>
				</button>

				<div class="auth-footer">
					Đã có tài khoản? <a href="#/login">Quay lại Đăng nhập</a>
				</div>
			</form>
		</div>
	`)
	khungChinh.ThemCon(theContainer)

	// Xử lý sự kiện đăng ký
	dom.LayPhanTuTheoId("btn-dang-ky").GanSuKien("click", func(this js.Value, args []js.Value) any {
		ten := strings.TrimSpace(dom.LayPhanTuTheoId("input-name").LayGiaTri())
		email := strings.TrimSpace(dom.LayPhanTuTheoId("input-email").LayGiaTri())
		matKhau := dom.LayPhanTuTheoId("input-password").LayGiaTri()
		xacNhan := dom.LayPhanTuTheoId("input-password-confirm").LayGiaTri()

		if ten == "" || email == "" || matKhau == "" {
			hienThiThongBaoAuth("error", "Vui lòng điền đầy đủ các thông tin bắt buộc.")
			return nil
		}

		if len(matKhau) < 6 {
			hienThiThongBaoAuth("error", "Mật khẩu phải có độ dài tối thiểu 6 ký tự.")
			return nil
		}

		if matKhau != xacNhan {
			hienThiThongBaoAuth("error", "Mật khẩu xác nhận không trùng khớp. Vui lòng kiểm tra lại.")
			return nil
		}

		nutDangKy := dom.LayPhanTuTheoId("btn-dang-ky")
		nutDangKy.DatThuocTinh("disabled", "true")
		nutDangKy.DatHtml(`<span class="spinner-sm"></span> Đang tạo tài khoản...`)

		go func() {
			err := api.DangKy(ten, email, matKhau, xacNhan)
			if err != nil {
				nutDangKy.XoaThuocTinh("disabled")
				nutDangKy.DatHtml(`<span class="btn-text">Tạo Tài Khoản</span>`)
				hienThiThongBaoAuth("error", fmt.Sprintf("❌ Đăng ký không thành công: %v", err))
			} else {
				hienThiThongBaoAuth("success", "✅ Tạo tài khoản thành công! Đang chuyển hướng...")
				routing.ChuyenHuong("#/dashboard")
			}
		}()
		return nil
	})
}
