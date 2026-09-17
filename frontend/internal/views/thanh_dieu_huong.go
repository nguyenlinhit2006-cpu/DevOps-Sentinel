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
 * Tạo thanh điều hướng (Navbar) chung cho toàn bộ các trang nội bộ của hệ thống.
 * Hỗ trợ làm nổi bật tab đang hoạt động và hiển thị vai trò người dùng tương ứng.
 */
func TaoThanhDieuHuong(tabHienTai string) dom.PhanTu {
	tt := state.LayTrangThai()
	tenNguoiDung := "Người dùng"
	vaiTro := "viewer"
	if tt.NguoiDungHienTai != nil {
		tenNguoiDung = tt.NguoiDungHienTai.Name
		vaiTro = tt.NguoiDungHienTai.Role
	}

	lopVaiTro := "badge-role-viewer"
	tenHienThiVaiTro := "Viewer"
	switch vaiTro {
	case "admin":
		lopVaiTro = "badge-role-admin"
		tenHienThiVaiTro = "Admin"
	case "team_lead":
		lopVaiTro = "badge-role-lead"
		tenHienThiVaiTro = "Team Lead"
	}

	lopTabDashboard := ""
	if tabHienTai == "dashboard" {
		lopTabDashboard = "active"
	}
	lopTabProjects := ""
	if tabHienTai == "projects" {
		lopTabProjects = "active"
	}
	lopTabAlerts := ""
	if tabHienTai == "alerts" {
		lopTabAlerts = "active"
	}
	lopTabTeams := ""
	if tabHienTai == "teams" {
		lopTabTeams = "active"
	}

	chuoiPolling := `<span class="pulse-dot-live"></span> Live (4s)`
	lopPolling := "badge-live"
	if !tt.PollingKichHoat {
		chuoiPolling = `⏸️ Tạm dừng`
		lopPolling = "badge-live paused"
	}

	thanhNav := dom.TaoPhanTu("header").ThemLopCss("app-navbar")
	thanhNav.DatHtml(fmt.Sprintf(`
		<div class="nav-left">
			<a href="#/dashboard" class="nav-brand">
				<span class="brand-radar">
					<span class="radar-dot"></span>
					<span class="radar-wave"></span>
				</span>
				<span class="brand-name">DevOps Sentinel</span>
				<span class="badge badge-wasm">WASM v1.0</span>
			</a>
			<nav class="nav-menu">
				<a href="#/dashboard" class="nav-link %s">
					<span class="icon">📊</span> Tổng quan
				</a>
				<a href="#/projects" class="nav-link %s">
					<span class="icon">📁</span> Dự án
				</a>
				<a href="#/alerts" class="nav-link %s">
					<span class="icon">🚨</span> Cảnh báo
				</a>
				<a href="#/teams" class="nav-link %s">
					<span class="icon">👥</span> Nhóm
				</a>
			</nav>
		</div>

		<div class="nav-user">
			<button id="btn-toggle-polling" class="badge %s" title="Nhấn để Bật/Tắt tự động cập nhật thời gian thực">
				%s
			</button>
			<div class="user-profile-chip">
				<div class="user-avatar">%s</div>
				<div class="user-meta">
					<span class="user-name">%s</span>
					<span class="badge %s">%s</span>
				</div>
			</div>
			<button id="btn-dang-xuat" class="btn btn-outline btn-sm btn-logout" title="Đăng xuất khỏi hệ thống">
				<span>🚪</span> Đăng xuất
			</button>
		</div>
	`, lopTabDashboard, lopTabProjects, lopTabAlerts, lopTabTeams,
		lopPolling, chuoiPolling,
		strings.ToUpper(string([]rune(tenNguoiDung)[0])),
		tenNguoiDung, lopVaiTro, tenHienThiVaiTro))

	// Đăng ký sự kiện toggle polling
	nutTogglePolling := thanhNav.Tim("#btn-toggle-polling")
	if nutTogglePolling.HopLe() {
		nutTogglePolling.GanSuKien("click", func(this js.Value, args []js.Value) any {
			kichHoat := tt.ChuyenDoiPolling()
			if kichHoat {
				nutTogglePolling.ThemLopCss("badge-live").XoaLopCss("paused")
				nutTogglePolling.DatHtml(`<span class="pulse-dot-live"></span> Live (4s)`)
			} else {
				nutTogglePolling.ThemLopCss("paused")
				nutTogglePolling.DatHtml(`⏸️ Tạm dừng`)
			}
			return nil
		})
	}

	// Đăng ký sự kiện click cho nút Đăng xuất
	nutDangXuat := thanhNav.Tim("#btn-dang-xuat")
	if nutDangXuat.HopLe() {
		nutDangXuat.GanSuKien("click", func(this js.Value, args []js.Value) any {
			go func() {
				_ = api.DangXuat()
				routing.ChuyenHuong("#/login")
			}()
			return nil
		})
	}

	return thanhNav
}
