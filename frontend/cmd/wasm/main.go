//go:build js && wasm

package main

import (
	"fmt"

	"devops-sentinel/frontend/internal/api"
	"devops-sentinel/frontend/internal/dom"
	"devops-sentinel/frontend/internal/routing"
	"devops-sentinel/frontend/internal/views"
)

/**
 * Điểm khởi nhập (Entrypoint) của ứng dụng Frontend Go WebAssembly.
 * Thiết lập DOM, State, API Client, đăng ký router và giữ cho tiến trình WASM luôn chạy.
 */
func main() {
	fmt.Println("🚀 [DevOps Sentinel Frontend] Go WebAssembly runtime đã khởi chạy thành công!")

	// Kiểm tra sức khỏe kết nối Backend ngầm
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

	// 1. Tuyến Đăng nhập
	bdh.DangKy("#/login", func(duongDan string, thamSo map[string]string) {
		views.HienThiTrangDangNhap()
	})

	// 2. Tuyến Đăng ký
	bdh.DangKy("#/register", func(duongDan string, thamSo map[string]string) {
		views.HienThiTrangDangKy()
	})

	// 3. Tuyến Bảng điều khiển (Dashboard)
	bdh.DangKy("#/dashboard", func(duongDan string, thamSo map[string]string) {
		views.HienThiTrangBangDieuKhien()
	})

	// 4. Tuyến Danh sách Quản lý Dự án
	bdh.DangKy("#/projects", func(duongDan string, thamSo map[string]string) {
		views.HienThiTrangDanhSachDuAn()
	})

	// 5. Tuyến Chi tiết Dự án (kèm ID động)
	bdh.DangKy("#/projects/{id}", func(duongDan string, thamSo map[string]string) {
		views.HienThiTrangChiTietDuAn(thamSo["id"])
	})

	// 6. Tuyến Chi tiết 1 Lượt chạy Pipeline (kèm ID động)
	bdh.DangKy("#/runs/{id}", func(duongDan string, thamSo map[string]string) {
		views.HienThiTrangChiTietLuotChay(thamSo["id"])
	})

	// Tuyến mặc định gốc "#"
	bdh.DangKy("#", func(duongDan string, thamSo map[string]string) {
		routing.ChuyenHuong("#/dashboard")
	})

	// Tuyến xử lý 404 khi không khớp đường dẫn
	bdh.DangKyKhongTimThay(func(duongDan string, thamSo map[string]string) {
		khungChinh := dom.LayPhanTuTheoId("app")
		khungChinh.XoaHetCon()

		theThongBao := dom.TaoPhanTu("div").ThemLopCss("thong-bao-404")
		theThongBao.DatHtml(`
			<div class="card-404 card">
				<h2>404 - Không tìm thấy trang</h2>
				<p>Đường dẫn bạn yêu cầu không tồn tại hoặc đã bị thay đổi.</p>
				<a href="#/dashboard" class="btn btn-primary" style="margin-top: 16px;">Về Bảng Điều Khiển</a>
			</div>
		`)
		khungChinh.ThemCon(theThongBao)
	})

	// Khởi chạy bộ định tuyến
	bdh.KhoiChay()

	// Giữ vòng lặp runtime WASM luôn hoạt động
	select {}
}
