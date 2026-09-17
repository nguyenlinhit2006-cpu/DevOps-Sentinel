//go:build js && wasm

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"syscall/js"

	"devops-sentinel/frontend/internal/state"
)

/**
 * Cấu trúc đóng gói chuẩn phản hồi JSON từ Backend RESTful API.
 */
type PhanHoiChuan struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
	Error   *ThongTinLoi    `json:"error"`
}

type ThongTinLoi struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

/**
 * Lớp khách hàng gọi API (API Client) sử dụng net/http tiêu chuẩn của Go trong môi trường WASM.
 */
type KhachHangApi struct {
	DuongDanGoc string
	HttpClient  *http.Client
}

var KhachHangDonNhat *KhachHangApi

/**
 * Khởi tạo hoặc lấy đối tượng API Client dùng chung.
 */
func LayKhachHangApi() *KhachHangApi {
	if KhachHangDonNhat == nil {
		// Mặc định gọi API cùng host hoặc trỏ tới backend port 8000 nếu dev
		duongDan := "/api/v1"
		cuaSo := js.Global().Get("window")
		if !cuaSo.IsNull() && !cuaSo.IsUndefined() {
			host := cuaSo.Get("location").Get("hostname").String()
			// Nếu frontend chạy cổng khác, có thể tự động trỏ về cổng backend 8000
			duongDan = fmt.Sprintf("http://%s:8000/api/v1", host)
		}

		KhachHangDonNhat = &KhachHangApi{
			DuongDanGoc: duongDan,
			HttpClient:  &http.Client{},
		}
	}
	return KhachHangDonNhat
}

/**
 * Thực hiện gửi HTTP request (GET, POST, PUT, DELETE) lên Backend.
 * Tự động gắn token xác thực và cơ chế Silent Refresh khi token hết hạn (mã lỗi 401).
 */
func (kh *KhachHangApi) GuiYeuCau(phuongThuc, duongDan string, duLieuGui any, canXacThuc bool) (*PhanHoiChuan, int, error) {
	toanBoDuongDan := fmt.Sprintf("%s%s", kh.DuongDanGoc, duongDan)

	var luongDoc io.Reader
	if duLieuGui != nil {
		chuoiJson, err := json.Marshal(duLieuGui)
		if err != nil {
			return nil, 0, fmt.Errorf("lỗi mã hóa json dữ liệu gửi: %w", err)
		}
		luongDoc = bytes.NewBuffer(chuoiJson)
	}

	yeuCau, err := http.NewRequest(phuongThuc, toanBoDuongDan, luongDoc)
	if err != nil {
		return nil, 0, fmt.Errorf("lỗi khởi tạo request: %w", err)
	}

	yeuCau.Header.Set("Content-Type", "application/json")
	yeuCau.Header.Set("Accept", "application/json")

	tt := state.LayTrangThai()
	if canXacThuc && tt.AccessToken != "" {
		yeuCau.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.AccessToken))
	}

	phanHoi, err := kh.HttpClient.Do(yeuCau)
	if err != nil {
		return nil, 0, fmt.Errorf("lỗi kết nối máy chủ API: %w", err)
	}
	defer phanHoi.Body.Close()

	// Xử lý Silent Refresh tự động khi gặp lỗi 401 Unauthorized
	if phanHoi.StatusCode == http.StatusUnauthorized && canXacThuc && tt.RefreshToken != "" {
		daLamMoiThanhCong := kh.ThucHienLamMoiToken()
		if daLamMoiThanhCong {
			// Gọi lại request ban đầu với Access Token mới
			return kh.GuiYeuCau(phuongThuc, duongDan, duLieuGui, canXacThuc)
		} else {
			tt.XoaPhienDangNhap()
			js.Global().Get("window").Get("location").Set("hash", "#/login")
			return nil, http.StatusUnauthorized, errors.New("phiên làm việc đã hết hạn, vui lòng đăng nhập lại")
		}
	}

	noiDungNhan, err := io.ReadAll(phanHoi.Body)
	if err != nil {
		return nil, phanHoi.StatusCode, fmt.Errorf("lỗi đọc nội dung phản hồi: %w", err)
	}

	var ketQua PhanHoiChuan
	if err := json.Unmarshal(noiDungNhan, &ketQua); err != nil {
		return nil, phanHoi.StatusCode, fmt.Errorf("lỗi giải mã json từ máy chủ: %w (Nội dung: %s)", err, string(noiDungNhan))
	}

	return &ketQua, phanHoi.StatusCode, nil
}

/**
 * Tự động gọi API Refresh Token khi nhận mã 401 để duy trì phiên làm việc mượt mà.
 */
func (kh *KhachHangApi) ThucHienLamMoiToken() bool {
	tt := state.LayTrangThai()
	if tt.RefreshToken == "" {
		return false
	}

	duLieuGui := map[string]string{"refresh_token": tt.RefreshToken}
	chuoiJson, _ := json.Marshal(duLieuGui)

	yeuCau, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/refresh", kh.DuongDanGoc), bytes.NewBuffer(chuoiJson))
	if err != nil {
		return false
	}
	yeuCau.Header.Set("Content-Type", "application/json")

	phanHoi, err := kh.HttpClient.Do(yeuCau)
	if err != nil || phanHoi.StatusCode != http.StatusOK {
		return false
	}
	defer phanHoi.Body.Close()

	noiDung, _ := io.ReadAll(phanHoi.Body)
	var ketQua struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}

	if err := json.Unmarshal(noiDung, &ketQua); err != nil || !ketQua.Success {
		return false
	}

	if tt.NguoiDungHienTai != nil {
		tt.LuuPhienDangNhap(ketQua.Data.AccessToken, ketQua.Data.RefreshToken, *tt.NguoiDungHienTai)
		return true
	}

	return false
}
