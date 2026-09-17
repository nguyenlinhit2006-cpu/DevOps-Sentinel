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
 * Hiển thị màn hình Quản lý Danh sách Toàn bộ Dự án (`#/projects`).
 * Cung cấp bộ lọc tìm kiếm theo thời gian thực và điều hướng tới chi tiết từng dự án.
 */
func HienThiTrangDanhSachDuAn() {
	if !state.LayTrangThai().DaDangNhap() {
		routing.ChuyenHuong("#/login")
		return
	}

	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	boKhung := dom.TaoPhanTu("div").ThemLopCss("dashboard-layout")
	thanhNav := TaoThanhDieuHuong("projects")
	boKhung.ThemCon(thanhNav)

	khungNoiDung := dom.TaoPhanTu("main").ThemLopCss("app-content")
	khungNoiDung.DatId("projects-page-content")

	khungNoiDung.DatHtml(`
		<div class="content-loading">
			<div class="spinner"></div>
			<h3>Đang tải danh sách dự án...</h3>
			<p>Vui lòng đợi giây lát trong khi WebAssembly tải dữ liệu từ máy chủ.</p>
		</div>
	`)
	boKhung.ThemCon(khungNoiDung)
	khungChinh.ThemCon(boKhung)

	taiDanhSachDuAnTrang(khungNoiDung)
}

/**
 * Nạp dữ liệu danh sách dự án từ Backend và render ra giao diện.
 */
func taiDanhSachDuAnTrang(khungNoiDung dom.PhanTu) {
	go func() {
		danhSach, meta, err := api.LayDanhSachDuAn("", 1, 50)
		if err != nil {
			khungNoiDung.DatHtml(fmt.Sprintf(`
				<div class="card card-error">
					<h3>❌ Không thể tải danh sách dự án</h3>
					<p>Lỗi kết nối: %v</p>
					<button id="btn-thu-lai-projects" class="btn btn-primary" style="margin-top: 16px;">
						🔄 Thử lại ngay
					</button>
				</div>
			`, err))

			nutThuLai := dom.LayPhanTuTheoId("btn-thu-lai-projects")
			if nutThuLai.HopLe() {
				nutThuLai.GanSuKien("click", func(this js.Value, args []js.Value) any {
					taiDanhSachDuAnTrang(khungNoiDung)
					return nil
				})
			}
			return
		}

		veGiaoDienDanhSachDuAn(khungNoiDung, danhSach, meta)
	}()
}

/**
 * Vẽ giao diện danh sách dự án chi tiết có ô tìm kiếm và các card dự án trực quan.
 */
func veGiaoDienDanhSachDuAn(khungNoiDung dom.PhanTu, danhSach []api.ThongTinDuAn, meta *api.PhanTrangMeta) {
	khungNoiDung.XoaHetCon()

	theHeader := dom.TaoPhanTu("div").ThemLopCss("page-header")
	theHeader.DatHtml(fmt.Sprintf(`
		<div class="header-left">
			<div class="breadcrumbs">
				<a href="#/dashboard" class="breadcrumb-item">Bảng điều khiển</a>
				<span class="breadcrumb-separator">/</span>
				<span class="breadcrumb-item active">Quản lý Dự án</span>
			</div>
			<h1 class="page-title">Danh Sách Dự Án CI/CD</h1>
			<p class="page-subtitle">Quản lý toàn bộ các kho lưu trữ đang kết nối và trạng thái giám sát tự động</p>
		</div>
		<div class="header-actions">
			<button id="btn-lam-moi-ds-du-an" class="btn btn-outline" title="Làm mới danh sách">
				<span class="icon">🔄</span> Làm mới
			</button>
		</div>
	`))
	khungNoiDung.ThemCon(theHeader)

	nutLamMoi := theHeader.Tim("#btn-lam-moi-ds-du-an")
	if nutLamMoi.HopLe() {
		nutLamMoi.GanSuKien("click", func(this js.Value, args []js.Value) any {
			nutLamMoi.DatHtml(`<span class="spinner-sm"></span> Đang tải...`)
			taiDanhSachDuAnTrang(khungNoiDung)
			return nil
		})
	}

	// Thanh tìm kiếm và bộ lọc dự án
	theBoLoc := dom.TaoPhanTu("div").ThemLopCss("card filter-toolbar")
	theBoLoc.DatHtml(`
		<div class="filter-search-group">
			<span class="search-icon">🔍</span>
			<input type="text" id="input-tim-du-an-trang" placeholder="Tìm theo tên dự án, đường dẫn slug hoặc mô tả..." />
		</div>
		<div class="filter-stats text-muted">
			Hiển thị <strong id="so-luong-hien-thi">0</strong> dự án
		</div>
	`)
	khungNoiDung.ThemCon(theBoLoc)

	// Lưới hiển thị các dự án
	khungLuoi := dom.TaoPhanTu("div").ThemLopCss("projects-grid").DatId("luoi-du-an-page")
	khungNoiDung.ThemCon(khungLuoi)

	veCardsDuAn := func(ds []api.ThongTinDuAn) {
		khungLuoi.XoaHetCon()

		theSoLuong := dom.LayPhanTuTheoId("so-luong-hien-thi")
		if theSoLuong.HopLe() {
			theSoLuong.DatNoiDung(fmt.Sprintf("%d", len(ds)))
		}

		if len(ds) == 0 {
			khungLuoi.DatHtml(`
				<div class="empty-state card" style="grid-column: 1 / -1;">
					<div class="empty-icon">📁</div>
					<h3>Không tìm thấy dự án</h3>
					<p>Không có dự án nào phù hợp với điều kiện tìm kiếm.</p>
				</div>
			`)
			return
		}

		for _, da := range ds {
			theCard := dom.TaoPhanTu("div").ThemLopCss("card project-card")

			thongTinRunHtml := `<div class="run-empty text-muted">Chưa ghi nhận lượt chạy nào</div>`
			if da.LuotChayMoiNhat != nil {
				tenTrangThai, lopBadge, iconTrangThai := DinhDangTrangThai(da.LuotChayMoiNhat.Status)
				hashNgan := da.LuotChayMoiNhat.CommitHash
				if len(hashNgan) > 7 {
					hashNgan = hashNgan[:7]
				}

				thongTinRunHtml = fmt.Sprintf(`
					<div class="latest-run-box">
						<div class="run-status-row">
							<span class="badge %s">%s %s</span>
							<span class="run-duration">⏱️ %s</span>
						</div>
						<div class="run-details-row">
							<span class="run-branch">🌿 %s</span>
							<span class="run-commit"><code>#%s</code></span>
						</div>
						<div class="run-time text-muted">%s</div>
					</div>
				`, lopBadge, iconTrangThai, tenTrangThai,
					DinhDangThoiLuong(da.LuotChayMoiNhat.DurationSeconds),
					da.LuotChayMoiNhat.Branch, hashNgan,
					DinhDangThoiGian(da.LuotChayMoiNhat.CreatedAt))
			}

			providerTen, providerLop, providerIcon := DinhDangNhaCungCap(da.CIProvider)

			theCanhBao := ""
			if da.ActiveAlertsCount > 0 {
				theCanhBao = fmt.Sprintf(`<span class="badge badge-danger-pulse">🚨 %d Cảnh báo</span>`, da.ActiveAlertsCount)
			}

			theCard.DatHtml(fmt.Sprintf(`
				<div class="project-header">
					<div class="project-meta-top">
						<span class="badge %s">%s %s</span>
						%s
					</div>
					<h3 class="project-name">%s</h3>
					<p class="project-desc text-muted">%s</p>
				</div>

				<div class="project-body">
					<div class="project-repo">
						<span class="icon">🔗</span>
						<a href="%s" target="_blank" rel="noopener noreferrer" class="repo-link text-truncate">
							%s
						</a>
					</div>
					<div class="project-latest-run">
						<div class="section-label">Lượt build gần nhất:</div>
						%s
					</div>
				</div>

				<div class="project-footer">
					<span class="text-muted" style="font-size: 0.8rem;">Ngưỡng lỗi: <strong>%d lần</strong></span>
					<a href="#/projects/%d" class="btn btn-primary btn-sm">
						Chi tiết dự án →
					</a>
				</div>
			`, providerLop, providerIcon, providerTen, theCanhBao,
				da.Name, da.Description,
				da.RepositoryURL, rutGonUrl(da.RepositoryURL),
				thongTinRunHtml,
				da.FailureThreshold, da.ID))

			khungLuoi.ThemCon(theCard)
		}
	}

	veCardsDuAn(danhSach)

	// Lắng nghe tìm kiếm thời gian thực
	oTimKiem := theBoLoc.Tim("#input-tim-du-an-trang")
	if oTimKiem.HopLe() {
		oTimKiem.GanSuKien("input", func(this js.Value, args []js.Value) any {
			tuKhoa := strings.ToLower(strings.TrimSpace(oTimKiem.LayGiaTri()))
			if tuKhoa == "" {
				veCardsDuAn(danhSach)
				return nil
			}

			dsLoc := make([]api.ThongTinDuAn, 0)
			for _, da := range danhSach {
				if strings.Contains(strings.ToLower(da.Name), tuKhoa) ||
					strings.Contains(strings.ToLower(da.Slug), tuKhoa) ||
					strings.Contains(strings.ToLower(da.Description), tuKhoa) {
					dsLoc = append(dsLoc, da)
				}
			}
			veCardsDuAn(dsLoc)
			return nil
		})
	}
}
