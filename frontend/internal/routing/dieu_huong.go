//go:build js && wasm

package routing

import (
	"strings"
	"syscall/js"

	"devops-sentinel/frontend/internal/dom"
	"devops-sentinel/frontend/internal/state"
)

/**
 * Kiểu định nghĩa hàm xử lý khi người dùng chuyển hướng tới một tuyến đường xác định.
 */
type HamXuLyTuyenDuong func(duongDan string, thamSo map[string]string)

/**
 * Bộ định tuyến điều hướng dựa trên Hash URL (#/dashboard, #/login...).
 */
type BoDieuHuong struct {
	CacTuyenDuong       map[string]HamXuLyTuyenDuong
	TuyenDuongMacDinh   string
	HamXuLyKhongTimThay HamXuLyTuyenDuong
}

var BoDieuHuongDonNhat *BoDieuHuong

/**
 * Khởi tạo hoặc lấy đối tượng Bộ điều hướng Singleton.
 */
func LayBoDieuHuong() *BoDieuHuong {
	if BoDieuHuongDonNhat == nil {
		BoDieuHuongDonNhat = &BoDieuHuong{
			CacTuyenDuong:     make(map[string]HamXuLyTuyenDuong),
			TuyenDuongMacDinh: "#/dashboard",
		}
	}
	return BoDieuHuongDonNhat
}

/**
 * Đăng ký một tuyến đường mới và hàm xử lý tương ứng.
 */
func (bdh *BoDieuHuong) DangKy(duongDan string, hamXuLy HamXuLyTuyenDuong) {
	bdh.CacTuyenDuong[duongDan] = hamXuLy
}

/**
 * Đăng ký hàm xử lý khi không tìm thấy tuyến đường (404 Not Found).
 */
func (bdh *BoDieuHuong) DangKyKhongTimThay(hamXuLy HamXuLyTuyenDuong) {
	bdh.HamXuLyKhongTimThay = hamXuLy
}

/**
 * Chuyển hướng người dùng sang một đường dẫn Hash mới.
 */
func ChuyenHuong(duongDanHash string) {
	if !strings.HasPrefix(duongDanHash, "#") {
		duongDanHash = "#" + duongDanHash
	}
	dom.LayCuaSo().GiaTri.Get("location").Set("hash", duongDanHash)
}

/**
 * Lấy đường dẫn hash hiện tại của trình duyệt (ví dụ: #/dashboard).
 */
func LayDuongDanHienTai() string {
	hash := dom.LayCuaSo().GiaTri.Get("location").Get("hash").String()
	if hash == "" || hash == "#" {
		return ""
	}
	return hash
}

/**
 * Khởi chạy bộ lắng nghe sự kiện thay đổi hash (hashchange) của trình duyệt.
 */
func (bdh *BoDieuHuong) KhoiChay() {
	cuaSo := dom.LayCuaSo()

	hamLangNghe := js.FuncOf(func(this js.Value, args []js.Value) any {
		bdh.DieuPhoi()
		return nil
	})

	cuaSo.GiaTri.Call("addEventListener", "hashchange", hamLangNghe)

	// Điều phối tuyến đường ngay khi vừa tải trang
	bdh.DieuPhoi()
}

/**
 * Phân tích URL hash hiện tại, kiểm tra đăng nhập và gọi hàm xử lý tương ứng.
 */
func (bdh *BoDieuHuong) DieuPhoi() {
	duongDan := LayDuongDanHienTai()
	tt := state.LayTrangThai()

	// Nếu chưa có hash, điều hướng về tuyến mặc định hoặc login
	if duongDan == "" {
		if tt.DaDangNhap() {
			ChuyenHuong(bdh.TuyenDuongMacDinh)
		} else {
			ChuyenHuong("#/login")
		}
		return
	}

	// Bảo vệ route: Nếu chưa đăng nhập mà cố vào các trang nghiệp vụ -> Chuyển về #/login
	if !tt.DaDangNhap() && duongDan != "#/login" && duongDan != "#/register" {
		ChuyenHuong("#/login")
		return
	}

	// Nếu đã đăng nhập mà cố vào #/login hoặc #/register -> Chuyển về Dashboard
	if tt.DaDangNhap() && (duongDan == "#/login" || duongDan == "#/register") {
		ChuyenHuong("#/dashboard")
		return
	}

	// 1. Tìm tuyến đường khớp chính xác
	if hamXuLy, tonTai := bdh.CacTuyenDuong[duongDan]; tonTai {
		hamXuLy(duongDan, make(map[string]string))
		return
	}

	// 2. Tìm tuyến đường có tham số động (ví dụ: #/projects/{id})
	for mauDuongDan, hamXuLy := range bdh.CacTuyenDuong {
		thamSo, khop := bdh.KiemTraKhopMau(mauDuongDan, duongDan)
		if khop {
			hamXuLy(duongDan, thamSo)
			return
		}
	}

	// 3. Nếu không tìm thấy, gọi hàm fallback hoặc về mặc định
	if bdh.HamXuLyKhongTimThay != nil {
		bdh.HamXuLyKhongTimThay(duongDan, nil)
	} else {
		ChuyenHuong(bdh.TuyenDuongMacDinh)
	}
}

/**
 * Kiểm tra xem hash hiện tại có khớp với mẫu có chứa tham số động (ví dụ: #/projects/{id}) không.
 */
func (bdh *BoDieuHuong) KiemTraKhopMau(mau, thucTe string) (map[string]string, bool) {
	cacPhanMau := strings.Split(strings.Trim(mau, "#/"), "/")
	cacPhanThucTe := strings.Split(strings.Trim(thucTe, "#/"), "/")

	if len(cacPhanMau) != len(cacPhanThucTe) {
		return nil, false
	}

	thamSo := make(map[string]string)
	for i := 0; i < len(cacPhanMau); i++ {
		if strings.HasPrefix(cacPhanMau[i], "{") && strings.HasSuffix(cacPhanMau[i], "}") {
			tenThamSo := strings.Trim(cacPhanMau[i], "{}")
			thamSo[tenThamSo] = cacPhanThucTe[i]
		} else if cacPhanMau[i] != cacPhanThucTe[i] {
			return nil, false
		}
	}

	return thamSo, true
}
