//go:build js && wasm

package api

import (
	"encoding/json"
	"errors"
	"fmt"

	"devops-sentinel/frontend/internal/state"
)

/**
 * Thực hiện gọi API đăng nhập tài khoản.
 */
func DangNhap(email, matKhau string) error {
	kh := LayKhachHangApi()
	duLieuGui := map[string]string{
		"email":    email,
		"password": matKhau,
	}

	phanHoi, _, err := kh.GuiYeuCau("POST", "/auth/login", duLieuGui, false)
	if err != nil {
		return err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return errors.New(phanHoi.Error.Message)
		}
		return errors.New("đăng nhập thất bại")
	}

	var ketQua struct {
		User         state.ThongTinNguoiDung `json:"user"`
		AccessToken  string                  `json:"access_token"`
		RefreshToken string                  `json:"refresh_token"`
	}

	if err := json.Unmarshal(phanHoi.Data, &ketQua); err != nil {
		return err
	}

	state.LayTrangThai().LuuPhienDangNhap(ketQua.AccessToken, ketQua.RefreshToken, ketQua.User)
	return nil
}

/**
 * Thực hiện gọi API đăng ký tài khoản mới.
 */
func DangKy(ten, email, matKhau, xacNhanMatKhau string) error {
	kh := LayKhachHangApi()
	duLieuGui := map[string]string{
		"name":                  ten,
		"email":                 email,
		"password":              matKhau,
		"password_confirmation": xacNhanMatKhau,
	}

	phanHoi, _, err := kh.GuiYeuCau("POST", "/auth/register", duLieuGui, false)
	if err != nil {
		return err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return errors.New(phanHoi.Error.Message)
		}
		return errors.New("đăng ký thất bại")
	}

	var ketQua struct {
		User         state.ThongTinNguoiDung `json:"user"`
		AccessToken  string                  `json:"access_token"`
		RefreshToken string                  `json:"refresh_token"`
	}

	if err := json.Unmarshal(phanHoi.Data, &ketQua); err != nil {
		return err
	}

	state.LayTrangThai().LuuPhienDangNhap(ketQua.AccessToken, ketQua.RefreshToken, ketQua.User)
	return nil
}

/**
 * Thực hiện gọi API đăng xuất và xóa dữ liệu phiên.
 */
func DangXuat() error {
	kh := LayKhachHangApi()
	kh.GuiYeuCau("POST", "/auth/logout", nil, true)
	state.LayTrangThai().XoaPhienDangNhap()
	return nil
}

/**
 * Lấy thông tin hồ sơ tài khoản hiện tại từ Backend.
 */
func LayThongTinCaNhan() (*state.ThongTinNguoiDung, error) {
	kh := LayKhachHangApi()
	phanHoi, _, err := kh.GuiYeuCau("GET", "/auth/me", nil, true)
	if err != nil {
		return nil, err
	}

	if !phanHoi.Success {
		return nil, errors.New("không thể lấy thông tin người dùng")
	}

	var ketQua struct {
		User state.ThongTinNguoiDung `json:"user"`
	}
	if err := json.Unmarshal(phanHoi.Data, &ketQua); err != nil {
		return nil, err
	}

	return &ketQua.User, nil
}

/**
 * Kiểm tra kết nối và trạng thái của hệ thống máy chủ Backend.
 */
func KiemTraSucKhoe() (string, error) {
	kh := LayKhachHangApi()
	phanHoi, _, err := kh.GuiYeuCau("GET", "/health", nil, false)
	if err != nil {
		return "", err
	}
	return phanHoi.Message, nil
}

/**
 * Lấy các chỉ số thống kê tóm tắt cho Bảng điều khiển (Dashboard).
 */
func LayThongKeBangDieuKhien() (*ThongKeBangDieuKhien, error) {
	kh := LayKhachHangApi()
	phanHoi, _, err := kh.GuiYeuCau("GET", "/dashboard/summary", nil, true)
	if err != nil {
		return nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, errors.New(phanHoi.Error.Message)
		}
		return nil, errors.New("không thể lấy thông tin thống kê bảng điều khiển")
	}

	var ketQua ThongKeBangDieuKhien
	if err := json.Unmarshal(phanHoi.Data, &ketQua); err != nil {
		return nil, err
	}

	return &ketQua, nil
}

/**
 * Lấy danh sách dự án kèm phân trang và tìm kiếm.
 */
func LayDanhSachDuAn(tuKhoa string, trang int, soLuong int) ([]ThongTinDuAn, *PhanTrangMeta, error) {
	kh := LayKhachHangApi()
	if trang <= 0 {
		trang = 1
	}
	if soLuong <= 0 {
		soLuong = 20
	}
	duongDan := fmt.Sprintf("/projects?page=%d&per_page=%d", trang, soLuong)
	if tuKhoa != "" {
		duongDan += fmt.Sprintf("&search=%s", tuKhoa)
	}

	phanHoi, _, err := kh.GuiYeuCau("GET", duongDan, nil, true)
	if err != nil {
		return nil, nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, nil, errors.New(phanHoi.Error.Message)
		}
		return nil, nil, errors.New("không thể lấy danh sách dự án")
	}

	var danhSach []ThongTinDuAn
	if err := json.Unmarshal(phanHoi.Data, &danhSach); err != nil {
		return nil, nil, err
	}

	return danhSach, phanHoi.Meta, nil
}

