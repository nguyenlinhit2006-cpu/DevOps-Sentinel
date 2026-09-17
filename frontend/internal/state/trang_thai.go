//go:build js && wasm

package state

import (
	"encoding/json"
	"syscall/js"
)

/**
 * Cấu trúc thông tin tóm tắt của người dùng đăng nhập.
 */
type ThongTinNguoiDung struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

/**
 * Trạng thái toàn cục của ứng dụng WebAssembly chạy trên client.
 */
type TrangThaiToanCuc struct {
	AccessToken      string
	RefreshToken     string
	NguoiDungHienTai *ThongTinNguoiDung
}

var trangThaiDonNhat *TrangThaiToanCuc

/**
 * Lấy đối tượng trạng thái toàn cục Singleton của ứng dụng.
 */
func LayTrangThai() *TrangThaiToanCuc {
	if trangThaiDonNhat == nil {
		trangThaiDonNhat = &TrangThaiToanCuc{}
		trangThaiDonNhat.TaiTuKhoLuuTru()
	}
	return trangThaiDonNhat
}

/**
 * Đọc dữ liệu phiên làm việc từ localStorage của trình duyệt khi nạp ứng dụng.
 */
func (tt *TrangThaiToanCuc) TaiTuKhoLuuTru() {
	khoLuuTru := js.Global().Get("localStorage")
	if khoLuuTru.IsNull() || khoLuuTru.IsUndefined() {
		return
	}

	tokenAccess := khoLuuTru.Call("getItem", "devops_sentinel_access_token")
	if !tokenAccess.IsNull() && !tokenAccess.IsUndefined() {
		tt.AccessToken = tokenAccess.String()
	}

	tokenRefresh := khoLuuTru.Call("getItem", "devops_sentinel_refresh_token")
	if !tokenRefresh.IsNull() && !tokenRefresh.IsUndefined() {
		tt.RefreshToken = tokenRefresh.String()
	}

	duLieuNguoiDung := khoLuuTru.Call("getItem", "devops_sentinel_user")
	if !duLieuNguoiDung.IsNull() && !duLieuNguoiDung.IsUndefined() {
		var nguoiDung ThongTinNguoiDung
		if err := json.Unmarshal([]byte(duLieuNguoiDung.String()), &nguoiDung); err == nil {
			tt.NguoiDungHienTai = &nguoiDung
		}
	}
}

/**
 * Lưu trữ phiên làm việc gồm token và thông tin người dùng vào localStorage.
 */
func (tt *TrangThaiToanCuc) LuuPhienDangNhap(tokenAccess, tokenRefresh string, nguoiDung ThongTinNguoiDung) {
	tt.AccessToken = tokenAccess
	tt.RefreshToken = tokenRefresh
	tt.NguoiDungHienTai = &nguoiDung

	khoLuuTru := js.Global().Get("localStorage")
	if khoLuuTru.IsNull() || khoLuuTru.IsUndefined() {
		return
	}

	khoLuuTru.Call("setItem", "devops_sentinel_access_token", tokenAccess)
	khoLuuTru.Call("setItem", "devops_sentinel_refresh_token", tokenRefresh)

	if chuoiJson, err := json.Marshal(nguoiDung); err == nil {
		khoLuuTru.Call("setItem", "devops_sentinel_user", string(chuoiJson))
	}
}

/**
 * Xóa sạch phiên làm việc trên bộ nhớ và trong localStorage khi đăng xuất.
 */
func (tt *TrangThaiToanCuc) XoaPhienDangNhap() {
	tt.AccessToken = ""
	tt.RefreshToken = ""
	tt.NguoiDungHienTai = nil

	khoLuuTru := js.Global().Get("localStorage")
	if !khoLuuTru.IsNull() && !khoLuuTru.IsUndefined() {
		khoLuuTru.Call("removeItem", "devops_sentinel_access_token")
		khoLuuTru.Call("removeItem", "devops_sentinel_refresh_token")
		khoLuuTru.Call("removeItem", "devops_sentinel_user")
	}
}

/**
 * Kiểm tra xem người dùng hiện tại đã đăng nhập và có Access Token hay chưa.
 */
func (tt *TrangThaiToanCuc) DaDangNhap() bool {
	return tt.AccessToken != "" && tt.NguoiDungHienTai != nil
}

/**
 * Kiểm tra xem người dùng có phải Quản trị viên hay không.
 */
func (tt *TrangThaiToanCuc) LaQuanTriVien() bool {
	return tt.NguoiDungHienTai != nil && tt.NguoiDungHienTai.Role == "admin"
}

/**
 * Kiểm tra xem người dùng có phải Trưởng nhóm trở lên hay không.
 */
func (tt *TrangThaiToanCuc) LaTruongNhom() bool {
	return tt.NguoiDungHienTai != nil && (tt.NguoiDungHienTai.Role == "admin" || tt.NguoiDungHienTai.Role == "team_lead")
}
