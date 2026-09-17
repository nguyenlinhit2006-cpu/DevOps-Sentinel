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
 * Hiển thị màn hình Bảng điều khiển tổng quan (Dashboard).
 * Tải số liệu tóm tắt và danh sách dự án từ Backend RESTful API và hiển thị trực quan.
 */
func HienThiTrangBangDieuKhien() {
	// Kiểm tra quyền truy cập (Route Guard)
	if !state.LayTrangThai().DaDangNhap() {
		routing.ChuyenHuong("#/login")
		return
	}

	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	boKhung := dom.TaoPhanTu("div").ThemLopCss("dashboard-layout")

	// 1. Gắn Thanh điều hướng Navbar
	thanhNav := TaoThanhDieuHuong("dashboard")
	boKhung.ThemCon(thanhNav)

	// 2. Khung nội dung chính
	khungNoiDung := dom.TaoPhanTu("main").ThemLopCss("app-content")
	khungNoiDung.DatId("dashboard-content")

	// Hiển thị trạng thái đang tải dữ liệu ban đầu
	khungNoiDung.DatHtml(`
		<div class="content-loading">
			<div class="spinner"></div>
			<h3>Đang đồng bộ dữ liệu giám sát...</h3>
			<p>Vui lòng đợi giây lát trong khi WebAssembly tải dữ liệu từ Backend API.</p>
		</div>
	`)
	boKhung.ThemCon(khungNoiDung)
	khungChinh.ThemCon(boKhung)

	// Tải dữ liệu từ Backend
	taiDuLieuVaHienThi(khungNoiDung)
}

/**
 * Gọi API backend để lấy thông tin thống kê và danh sách dự án, sau đó vẽ lại giao diện.
 */
func taiDuLieuVaHienThi(khungNoiDung dom.PhanTu) {
	go func() {
		thongKe, errThongKe := api.LayThongKeBangDieuKhien()
		danhSachDuAn, _, errDuAn := api.LayDanhSachDuAn("", 1, 30)

		if errThongKe != nil {
			khungNoiDung.DatHtml(fmt.Sprintf(`
				<div class="card card-error">
					<h3>❌ Không thể tải dữ liệu Dashboard</h3>
					<p>Lỗi kết nối: %v</p>
					<button id="btn-thu-lai" class="btn btn-primary" style="margin-top: 16px;">
						🔄 Thử lại ngay
					</button>
				</div>
			`, errThongKe))

			nutThuLai := dom.LayPhanTuTheoId("btn-thu-lai")
			if nutThuLai.HopLe() {
				nutThuLai.GanSuKien("click", func(this js.Value, args []js.Value) any {
					taiDuLieuVaHienThi(khungNoiDung)
					return nil
				})
			}
			return
		}

		// Vẽ giao diện Dashboard hoàn chỉnh khi có dữ liệu
		veGiaoDienTongQuan(khungNoiDung, thongKe, danhSachDuAn, errDuAn)
	}()
}

/**
 * Vẽ toàn bộ các thành phần của Dashboard (Tiêu đề, 4 thẻ KPI, Lưới dự án, Bảng hoạt động).
 */
func veGiaoDienTongQuan(khungNoiDung dom.PhanTu, tk *api.ThongKeBangDieuKhien, danhSachDuAn []api.ThongTinDuAn, errDuAn error) {
	khungNoiDung.XoaHetCon()

	// 1. Header trang Dashboard
	theHeader := dom.TaoPhanTu("div").ThemLopCss("page-header")
	theHeader.DatHtml(`
		<div class="header-left">
			<h1 class="page-title">Tổng Quan Hệ Thống</h1>
			<p class="page-subtitle">Giám sát các đường ống CI/CD, tỷ lệ thành công và cảnh báo thời gian thực</p>
		</div>
		<div class="header-actions">
			<button id="btn-lam-moi" class="btn btn-outline" title="Làm mới dữ liệu từ máy chủ">
				<span class="icon">🔄</span> Làm mới
			</button>
		</div>
	`)
	khungNoiDung.ThemCon(theHeader)

	// Gắn sự kiện nút làm mới
	nutLamMoi := theHeader.Tim("#btn-lam-moi")
	if nutLamMoi.HopLe() {
		nutLamMoi.GanSuKien("click", func(this js.Value, args []js.Value) any {
			nutLamMoi.DatHtml(`<span class="spinner-sm"></span> Đang tải...`)
			taiDuLieuVaHienThi(khungNoiDung)
			return nil
		})
	}

	// 2. Bốn Thẻ Chỉ số KPI Tổng quan (KPI Metrics Cards)
	theKpiContainer := dom.TaoPhanTu("div").ThemLopCss("kpi-grid")

	// Xác định màu sắc và biểu tượng cho tỷ lệ thành công
	mauTyLe := "rate-high"
	if tk.SuccessRatePercent < 70 {
		mauTyLe = "rate-low"
	} else if tk.SuccessRatePercent < 90 {
		mauTyLe = "rate-medium"
	}

	// Xác định trạng thái cảnh báo
	theCanhBaoHtml := `<span class="badge badge-success-subtle">Hệ thống ổn định</span>`
	lopCanhBaoCard := ""
	if tk.ActiveAlertsCount > 0 {
		theCanhBaoHtml = fmt.Sprintf(`<span class="badge badge-danger-subtle"><span class="pulse-dot"></span> %d Cảnh báo đang mở</span>`, tk.ActiveAlertsCount)
		lopCanhBaoCard = "card-alert-danger"
	}

	theKpiContainer.DatHtml(fmt.Sprintf(`
		<!-- Thẻ 1: Tổng số dự án -->
		<div class="card kpi-card">
			<div class="kpi-header">
				<span class="kpi-label">Tổng số Dự án</span>
				<div class="kpi-icon-wrap icon-projects">📁</div>
			</div>
			<div class="kpi-value">%d</div>
			<div class="kpi-footer">
				<span class="text-muted">Dự án được giám sát liên tục</span>
			</div>
		</div>

		<!-- Thẻ 2: Lượt chạy hôm nay -->
		<div class="card kpi-card">
			<div class="kpi-header">
				<span class="kpi-label">Lượt chạy hôm nay</span>
				<div class="kpi-icon-wrap icon-runs">🚀</div>
			</div>
			<div class="kpi-value">%d</div>
			<div class="kpi-footer">
				<span class="text-muted">Tổng tích lũy: <strong>%d</strong> lượt build</span>
			</div>
		</div>

		<!-- Thẻ 3: Tỷ lệ thành công -->
		<div class="card kpi-card">
			<div class="kpi-header">
				<span class="kpi-label">Tỷ lệ Thành công</span>
				<div class="kpi-icon-wrap icon-rate">📈</div>
			</div>
			<div class="kpi-value %s">%.1f%%</div>
			<div class="kpi-footer">
				<div class="progress-bar-bg">
					<div class="progress-bar-fill %s" style="width: %.1f%%;"></div>
				</div>
			</div>
		</div>

		<!-- Thẻ 4: Cảnh báo Sentinel -->
		<div class="card kpi-card %s">
			<div class="kpi-header">
				<span class="kpi-label">Cảnh báo Sentinel</span>
				<div class="kpi-icon-wrap icon-alerts">🚨</div>
			</div>
			<div class="kpi-value">%d</div>
			<div class="kpi-footer">
				%s
			</div>
		</div>
	`, tk.TotalProjects, tk.TotalRunsToday, tk.TotalRuns,
		mauTyLe, tk.SuccessRatePercent, mauTyLe, tk.SuccessRatePercent,
		lopCanhBaoCard, tk.ActiveAlertsCount, theCanhBaoHtml))
	khungNoiDung.ThemCon(theKpiContainer)

	// 3. Khung phân chia 2 phân vùng chính (Lưới dự án & Hoạt động gần đây)
	khungBoCuc := dom.TaoPhanTu("div").ThemLopCss("dashboard-sections")

	// Phân vùng A: Danh sách Dự án
	phanVungDuAn := dom.TaoPhanTu("section").ThemLopCss("section-projects")
	vePhanVungDuAn(phanVungDuAn, danhSachDuAn, errDuAn)
	khungBoCuc.ThemCon(phanVungDuAn)

	// Phân vùng B: Hoạt động Pipeline Gần nhất
	phanVungHoatDong := dom.TaoPhanTu("section").ThemLopCss("section-activity")
	vePhanVungHoatDong(phanVungHoatDong, tk.RecentRuns)
	khungBoCuc.ThemCon(phanVungHoatDong)

	khungNoiDung.ThemCon(khungBoCuc)
}

/**
 * Vẽ phân vùng danh sách dự án kèm ô tìm kiếm lọc nhanh theo tên.
 */
func vePhanVungDuAn(phanVung dom.PhanTu, danhSachDuAn []api.ThongTinDuAn, errDuAn error) {
	soLuong := len(danhSachDuAn)
	headerHtml := fmt.Sprintf(`
		<div class="section-header">
			<div class="section-title-wrap">
				<h2 class="section-title">Dự Án Đang Giám Sát</h2>
				<span class="badge badge-count">%d dự án</span>
			</div>
			<div class="project-search-box">
				<span class="search-icon">🔍</span>
				<input type="text" id="input-loc-du-an" placeholder="Tìm kiếm theo tên dự án..." />
			</div>
		</div>
	`, soLuong)
	phanVung.DatHtml(headerHtml)

	khungLuoi := dom.TaoPhanTu("div").ThemLopCss("projects-grid").DatId("luoi-du-an")

	if errDuAn != nil {
		khungLuoi.DatHtml(fmt.Sprintf(`<div class="alert-box alert-error">Lỗi nạp danh sách dự án: %v</div>`, errDuAn))
		phanVung.ThemCon(khungLuoi)
		return
	}

	if soLuong == 0 {
		khungLuoi.DatHtml(`
			<div class="empty-state card">
				<div class="empty-icon">📁</div>
				<h3>Chưa có dự án nào</h3>
				<p>Hệ thống chưa có dự án nào được đăng ký theo dõi CI/CD.</p>
			</div>
		`)
		phanVung.ThemCon(khungLuoi)
		return
	}

	// Hàm render danh sách card dự án
	veDanhSachTheDuAn := func(ds []api.ThongTinDuAn) {
		khungLuoi.XoaHetCon()
		if len(ds) == 0 {
			khungLuoi.DatHtml(`
				<div class="empty-state card" style="grid-column: 1 / -1;">
					<p>Không tìm thấy dự án nào khớp với từ khóa tìm kiếm.</p>
				</div>
			`)
			return
		}

		for _, da := range ds {
			theDuAn := dom.TaoPhanTu("div").ThemLopCss("card project-card")

			// Xử lý thông tin lượt chạy gần nhất
			thongTinRunHtml := `<div class="run-empty text-muted">Chưa có lượt build nào</div>`
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
							<span class="run-branch" title="Nhánh">🌿 %s</span>
							<span class="run-commit" title="Commit Hash"><code>#%s</code></span>
						</div>
						<div class="run-time text-muted">%s</div>
					</div>
				`, lopBadge, iconTrangThai, tenTrangThai,
					DinhDangThoiLuong(da.LuotChayMoiNhat.DurationSeconds),
					da.LuotChayMoiNhat.Branch, hashNgan,
					DinhDangThoiGian(da.LuotChayMoiNhat.CreatedAt))
			}

			// Huy hiệu Provider
			providerTen, providerLop, providerIcon := DinhDangNhaCungCap(da.CIProvider)

			// Huy hiệu cảnh báo nếu có
			theCanhBaoDuAn := ""
			if da.ActiveAlertsCount > 0 {
				theCanhBaoDuAn = fmt.Sprintf(`<span class="badge badge-danger-pulse" title="Dự án có cảnh báo fail liên tiếp">🚨 %d Cảnh báo</span>`, da.ActiveAlertsCount)
			}

			theDuAn.DatHtml(fmt.Sprintf(`
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
					<a href="#/projects/%d" class="btn btn-outline btn-sm">
						Xem chi tiết →
					</a>
				</div>
			`, providerLop, providerIcon, providerTen, theCanhBaoDuAn,
				da.Name, da.Description,
				da.RepositoryURL, rutGonUrl(da.RepositoryURL),
				thongTinRunHtml,
				da.FailureThreshold, da.ID))

			khungLuoi.ThemCon(theDuAn)
		}
	}

	veDanhSachTheDuAn(danhSachDuAn)
	phanVung.ThemCon(khungLuoi)

	// Gắn sự kiện tìm kiếm lọc danh sách thời gian thực
	oTimKiem := phanVung.Tim("#input-loc-du-an")
	if oTimKiem.HopLe() {
		oTimKiem.GanSuKien("input", func(this js.Value, args []js.Value) any {
			tuKhoa := strings.ToLower(strings.TrimSpace(oTimKiem.LayGiaTri()))
			if tuKhoa == "" {
				veDanhSachTheDuAn(danhSachDuAn)
				return nil
			}

			danhSachLoc := make([]api.ThongTinDuAn, 0)
			for _, da := range danhSachDuAn {
				if strings.Contains(strings.ToLower(da.Name), tuKhoa) ||
					strings.Contains(strings.ToLower(da.Slug), tuKhoa) ||
					strings.Contains(strings.ToLower(da.Description), tuKhoa) {
					danhSachLoc = append(danhSachLoc, da)
				}
			}
			veDanhSachTheDuAn(danhSachLoc)
			return nil
		})
	}
}

/**
 * Vẽ phân vùng hiển thị 8 lượt chạy gần đây nhất trên toàn hệ thống.
 */
func vePhanVungHoatDong(phanVung dom.PhanTu, danhSachChay []api.ThongTinLuotChay) {
	headerHtml := `
		<div class="section-header">
			<div class="section-title-wrap">
				<h2 class="section-title">Hoạt Động Pipeline Gần Nhất</h2>
				<span class="badge badge-count">Thời gian thực</span>
			</div>
		</div>
	`
	phanVung.DatHtml(headerHtml)

	theCard := dom.TaoPhanTu("div").ThemLopCss("card activity-card")

	if len(danhSachChay) == 0 {
		theCard.DatHtml(`
			<div class="empty-state">
				<p class="text-muted">Chưa ghi nhận hoạt động pipeline nào trong hệ thống.</p>
			</div>
		`)
		phanVung.ThemCon(theCard)
		return
	}

	theDanhSach := dom.TaoPhanTu("div").ThemLopCss("activity-list")

	for _, lc := range danhSachChay {
		theMuc := dom.TaoPhanTu("div").ThemLopCss("activity-item")
		tenTrangThai, lopBadge, iconTrangThai := DinhDangTrangThai(lc.Status)

		tenDuAn := "Dự án"
		if lc.DuAn != nil {
			tenDuAn = lc.DuAn.Name
		}

		hashNgan := lc.CommitHash
		if len(hashNgan) > 7 {
			hashNgan = hashNgan[:7]
		}

		thongDiepCommit := lc.CommitMessage
		if thongDiepCommit == "" {
			thongDiepCommit = "Không có mô tả commit"
		}
		if len(thongDiepCommit) > 55 {
			thongDiepCommit = thongDiepCommit[:52] + "..."
		}

		theMuc.DatHtml(fmt.Sprintf(`
			<div class="activity-status">
				<span class="badge %s" title="%s">%s %s</span>
			</div>
			<div class="activity-info">
				<div class="activity-title-row">
					<strong class="activity-project-name">%s</strong>
					<span class="activity-pipeline text-muted">/ %s</span>
					<span class="activity-branch">🌿 %s</span>
				</div>
				<div class="activity-commit-row">
					<code class="commit-hash">#%s</code>
					<span class="commit-message text-muted">%s</span>
				</div>
			</div>
			<div class="activity-meta">
				<span class="activity-author">👤 %s</span>
				<span class="activity-time text-muted">⏱️ %s • %s</span>
			</div>
		`, lopBadge, tenTrangThai, iconTrangThai, tenTrangThai,
			tenDuAn, lc.PipelineName, lc.Branch,
			hashNgan, thongDiepCommit,
			lc.Author, DinhDangThoiLuong(lc.DurationSeconds),
			DinhDangThoiGian(lc.CreatedAt)))

		theDanhSach.ThemCon(theMuc)
	}

	theCard.ThemCon(theDanhSach)
	phanVung.ThemCon(theCard)
}

/**
 * Định dạng chuỗi trạng thái thành Tên tiếng Việt, Lớp CSS Badge và Icon tương ứng.
 */
func DinhDangTrangThai(trangThai string) (tenHienThi, lopBadge, icon string) {
	switch strings.ToLower(trangThai) {
	case "success":
		return "Thành công", "badge-success", "✓"
	case "failed":
		return "Thất bại", "badge-failed", "✗"
	case "running":
		return "Đang chạy", "badge-running", "⚡"
	case "queued":
		return "Đang đợi", "badge-queued", "⏳"
	case "cancelled":
		return "Đã hủy", "badge-cancelled", "⊘"
	default:
		return trangThai, "badge-neutral", "•"
	}
}

/**
 * Định dạng thông tin nhà cung cấp CI/CD.
 */
func DinhDangNhaCungCap(provider string) (tenHienThi, lopBadge, icon string) {
	switch strings.ToLower(provider) {
	case "github":
		return "GitHub Actions", "badge-provider-github", "🐙"
	case "gitlab":
		return "GitLab CI", "badge-provider-gitlab", "🦊"
	default:
		return "Generic CI", "badge-provider-generic", "⚙️"
	}
}

/**
 * Định dạng thời lượng chạy tính theo giây thành chuỗi dễ đọc (VD: 45s, 2m 15s).
 */
func DinhDangThoiLuong(giay int) string {
	if giay <= 0 {
		return "0s"
	}
	if giay < 60 {
		return fmt.Sprintf("%ds", giay)
	}
	phut := giay / 60
	giayConLai := giay % 60
	if giayConLai == 0 {
		return fmt.Sprintf("%dm", phut)
	}
	return fmt.Sprintf("%dm %ds", phut, giayConLai)
}

/**
 * Định dạng mốc thời gian ISO thành định dạng ngắn gọn hiển thị.
 */
func DinhDangThoiGian(isoStr string) string {
	if isoStr == "" {
		return "-"
	}
	// Cắt ngắn chuỗi ISO: 2026-09-17T06:00:00Z -> 06:00 17/09
	if len(isoStr) >= 16 {
		ngayThang := isoStr[:10]
		gioPhut := isoStr[11:16]
		return fmt.Sprintf("%s %s", gioPhut, ngayThang)
	}
	return isoStr
}

/**
 * Rút gọn đường dẫn URL kho lưu trữ để hiển thị thẩm mỹ trên thẻ.
 */
func rutGonUrl(url string) string {
	rutGon := strings.TrimPrefix(url, "https://")
	rutGon = strings.TrimPrefix(rutGon, "http://")
	if len(rutGon) > 35 {
		return rutGon[:32] + "..."
	}
	return rutGon
}
