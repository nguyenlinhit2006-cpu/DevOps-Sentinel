//go:build js && wasm

package api

import (
	"encoding/json"
	"errors"

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
