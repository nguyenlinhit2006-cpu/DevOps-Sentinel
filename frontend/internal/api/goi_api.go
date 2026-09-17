//go:build js && wasm

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

/**
 * Lấy thông tin chi tiết của một dự án kèm thông tin cấu hình webhook.
 */
func LayChiTietDuAn(id int64) (*ThongTinDuAnChiTiet, error) {
	kh := LayKhachHangApi()
	duongDan := fmt.Sprintf("/projects/%d", id)

	phanHoi, _, err := kh.GuiYeuCau("GET", duongDan, nil, true)
	if err != nil {
		return nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, errors.New(phanHoi.Error.Message)
		}
		return nil, errors.New("không tìm thấy dự án yêu cầu")
	}

	var ketQua ThongTinDuAnChiTiet
	if err := json.Unmarshal(phanHoi.Data, &ketQua); err != nil {
		return nil, err
	}

	return &ketQua, nil
}

/**
 * Lấy danh sách lịch sử lượt chạy pipeline của một dự án có phân trang và bộ lọc.
 */
func LayLichSuLuotChayDuAn(duAnId int64, trang int, soLuong int, nhanh string, trangThai string) ([]ThongTinLuotChay, *PhanTrangMeta, error) {
	kh := LayKhachHangApi()
	if trang <= 0 {
		trang = 1
	}
	if soLuong <= 0 {
		soLuong = 10
	}

	duongDan := fmt.Sprintf("/projects/%d/runs?page=%d&per_page=%d", duAnId, trang, soLuong)
	if nhanh != "" {
		duongDan += fmt.Sprintf("&branch=%s", nhanh)
	}
	if trangThai != "" && trangThai != "all" {
		duongDan += fmt.Sprintf("&status=%s", trangThai)
	}

	phanHoi, _, err := kh.GuiYeuCau("GET", duongDan, nil, true)
	if err != nil {
		return nil, nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, nil, errors.New(phanHoi.Error.Message)
		}
		return nil, nil, errors.New("không thể lấy lịch sử lượt chạy")
	}

	var danhSach []ThongTinLuotChay
	if err := json.Unmarshal(phanHoi.Data, &danhSach); err != nil {
		return nil, nil, err
	}

	return danhSach, phanHoi.Meta, nil
}

/**
 * Lấy thông tin chi tiết của một lượt chạy cụ thể (Bao gồm log tóm tắt).
 */
func LayChiTietLuotChay(id int64) (*ThongTinLuotChay, error) {
	kh := LayKhachHangApi()
	duongDan := fmt.Sprintf("/runs/%d", id)

	phanHoi, _, err := kh.GuiYeuCau("GET", duongDan, nil, true)
	if err != nil {
		return nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, errors.New(phanHoi.Error.Message)
		}
		return nil, errors.New("không tìm thấy lượt chạy yêu cầu")
	}

	var ketQua ThongTinLuotChay
	if err := json.Unmarshal(phanHoi.Data, &ketQua); err != nil {
		return nil, err
	}

	return &ketQua, nil
}

/**
 * Yêu cầu máy chủ sinh lại khóa bí mật Webhook Secret mới cho dự án.
 */
func TaoLaiWebhookSecret(duAnId int64) (string, error) {
	kh := LayKhachHangApi()
	duongDan := fmt.Sprintf("/projects/%d/regenerate-secret", duAnId)

	phanHoi, _, err := kh.GuiYeuCau("POST", duongDan, nil, true)
	if err != nil {
		return "", err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return "", errors.New(phanHoi.Error.Message)
		}
		return "", errors.New("không thể sinh lại webhook secret")
	}

	var ketQua struct {
		WebhookSecret string `json:"webhook_secret"`
	}
	if err := json.Unmarshal(phanHoi.Data, &ketQua); err != nil {
		return "", err
	}

	return ketQua.WebhookSecret, nil
}

/**
 * Lấy danh sách các cảnh báo sự cố theo trạng thái (active/resolved) và/hoặc dự án.
 */
func LayDanhSachCanhBao(trangThai string, duAnId int64) ([]ThongTinCanhBao, error) {
	kh := LayKhachHangApi()
	duongDan := "/alerts"

	thamSo := make([]string, 0)
	if trangThai != "" && trangThai != "all" {
		thamSo = append(thamSo, fmt.Sprintf("status=%s", trangThai))
	}
	if duAnId > 0 {
		thamSo = append(thamSo, fmt.Sprintf("project_id=%d", duAnId))
	}

	if len(thamSo) > 0 {
		duongDan += "?" + strings.Join(thamSo, "&")
	}

	phanHoi, _, err := kh.GuiYeuCau("GET", duongDan, nil, true)
	if err != nil {
		return nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, errors.New(phanHoi.Error.Message)
		}
		return nil, errors.New("không thể lấy danh sách cảnh báo")
	}

	var danhSach []ThongTinCanhBao
	if err := json.Unmarshal(phanHoi.Data, &danhSach); err != nil {
		return nil, err
	}

	return danhSach, nil
}

/**
 * Xác nhận đóng và khắc phục một cảnh báo sự cố (Yêu cầu vai trò Team Lead trở lên).
 */
func DongCanhBao(id int64) error {
	kh := LayKhachHangApi()
	duongDan := fmt.Sprintf("/alerts/%d/resolve", id)

	phanHoi, _, err := kh.GuiYeuCau("POST", duongDan, nil, true)
	if err != nil {
		return err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return errors.New(phanHoi.Error.Message)
		}
		return errors.New("không thể đóng cảnh báo")
	}

	return nil
}

/**
 * Lấy danh sách tất cả các nhóm làm việc trong hệ thống.
 */
func LayDanhSachNhom() ([]ThongTinNhom, error) {
	kh := LayKhachHangApi()
	phanHoi, _, err := kh.GuiYeuCau("GET", "/teams", nil, true)
	if err != nil {
		return nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, errors.New(phanHoi.Error.Message)
		}
		return nil, errors.New("không thể lấy danh sách nhóm")
	}

	var danhSach []ThongTinNhom
	if err := json.Unmarshal(phanHoi.Data, &danhSach); err != nil {
		return nil, err
	}

	return danhSach, nil
}

/**
 * Xem thông tin chi tiết của một nhóm làm việc kèm danh sách thành viên và dự án trực thuộc.
 */
func LayChiTietNhom(id int64) (*ThongTinNhomChiTiet, error) {
	kh := LayKhachHangApi()
	duongDan := fmt.Sprintf("/teams/%d", id)

	phanHoi, _, err := kh.GuiYeuCau("GET", duongDan, nil, true)
	if err != nil {
		return nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, errors.New(phanHoi.Error.Message)
		}
		return nil, errors.New("không tìm thấy nhóm yêu cầu")
	}

	var ketQua ThongTinNhomChiTiet
	if err := json.Unmarshal(phanHoi.Data, &ketQua); err != nil {
		return nil, err
	}

	return &ketQua, nil
}

/**
 * Tạo một nhóm làm việc mới (Yêu cầu vai trò Admin).
 */
func TaoNhomMoi(ten, moTa string) (*ThongTinNhom, error) {
	kh := LayKhachHangApi()
	duLieuGui := map[string]string{
		"name":        ten,
		"description": moTa,
	}

	phanHoi, _, err := kh.GuiYeuCau("POST", "/teams", duLieuGui, true)
	if err != nil {
		return nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, errors.New(phanHoi.Error.Message)
		}
		return nil, errors.New("không thể tạo nhóm mới")
	}

	var nhomMoi ThongTinNhom
	if err := json.Unmarshal(phanHoi.Data, &nhomMoi); err != nil {
		return nil, err
	}

	return &nhomMoi, nil
}

/**
 * Thêm hoặc gán vai trò thành viên vào một nhóm làm việc.
 */
func ThemThanhVienVaoNhom(nhomId int64, userId int64, vaiTro string) error {
	kh := LayKhachHangApi()
	duongDan := fmt.Sprintf("/teams/%d/members", nhomId)
	duLieuGui := map[string]any{
		"user_id": userId,
		"role":    vaiTro,
	}

	phanHoi, _, err := kh.GuiYeuCau("POST", duongDan, duLieuGui, true)
	if err != nil {
		return err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return errors.New(phanHoi.Error.Message)
		}
		return errors.New("không thể thêm thành viên vào nhóm")
	}

	return nil
}

/**
 * Xóa một thành viên ra khỏi nhóm làm việc.
 */
func XoaThanhVienKhoiNhom(nhomId int64, userId int64) error {
	kh := LayKhachHangApi()
	duongDan := fmt.Sprintf("/teams/%d/members/%d", nhomId, userId)

	phanHoi, _, err := kh.GuiYeuCau("DELETE", duongDan, nil, true)
	if err != nil {
		return err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return errors.New(phanHoi.Error.Message)
		}
		return errors.New("không thể xóa thành viên khỏi nhóm")
	}

	return nil
}

/**
 * Xóa toàn bộ nhóm làm việc (Yêu cầu vai trò Admin).
 */
func XoaNhom(id int64) error {
	kh := LayKhachHangApi()
	duongDan := fmt.Sprintf("/teams/%d", id)

	phanHoi, _, err := kh.GuiYeuCau("DELETE", duongDan, nil, true)
	if err != nil {
		return err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return errors.New(phanHoi.Error.Message)
		}
		return errors.New("không thể xóa nhóm làm việc")
	}

	return nil
}

/**
 * Lấy danh sách tất cả người dùng trong hệ thống (dùng để chọn thành viên đưa vào nhóm).
 */
func LayDanhSachNguoiDung() ([]ThongTinThanhVien, error) {
	kh := LayKhachHangApi()
	phanHoi, _, err := kh.GuiYeuCau("GET", "/users", nil, true)
	if err != nil {
		return nil, err
	}

	if !phanHoi.Success {
		if phanHoi.Error != nil {
			return nil, errors.New(phanHoi.Error.Message)
		}
		return nil, errors.New("không thể lấy danh sách người dùng")
	}

	var danhSach []ThongTinThanhVien
	if err := json.Unmarshal(phanHoi.Data, &danhSach); err != nil {
		return nil, err
	}

	return danhSach, nil
}




