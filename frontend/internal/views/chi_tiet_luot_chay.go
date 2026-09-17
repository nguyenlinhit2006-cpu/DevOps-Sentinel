//go:build js && wasm

package views

import (
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"devops-sentinel/frontend/internal/api"
	"devops-sentinel/frontend/internal/dom"
	"devops-sentinel/frontend/internal/routing"
	"devops-sentinel/frontend/internal/state"
)

/**
 * Hiển thị màn hình Chi tiết 1 Lượt chạy Pipeline (`#/runs/{id}`).
 * Bao gồm thông tin commit, nhánh, tác giả, thời lượng và trình xem Log Monospace Terminal (tối đa 64KB).
 */
func HienThiTrangChiTietLuotChay(idChuoi string) {
	if !state.LayTrangThai().DaDangNhap() {
		routing.ChuyenHuong("#/login")
		return
	}

	id, err := strconv.ParseInt(idChuoi, 10, 64)
	if err != nil || id <= 0 {
		routing.ChuyenHuong("#/dashboard")
		return
	}

	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	boKhung := dom.TaoPhanTu("div").ThemLopCss("dashboard-layout")
	thanhNav := TaoThanhDieuHuong("projects")
	boKhung.ThemCon(thanhNav)

	khungNoiDung := dom.TaoPhanTu("main").ThemLopCss("app-content")
	khungNoiDung.DatId("run-detail-content")

	khungNoiDung.DatHtml(`
		<div class="content-loading">
			<div class="spinner"></div>
			<h3>Đang tải chi tiết lượt chạy pipeline...</h3>
			<p>Vui lòng đợi giây lát trong khi WebAssembly tải dữ liệu.</p>
		</div>
	`)
	boKhung.ThemCon(khungNoiDung)
	khungChinh.ThemCon(boKhung)

	go func() {
		run, errRun := api.LayChiTietLuotChay(id)
		if errRun != nil {
			khungNoiDung.DatHtml(fmt.Sprintf(`
				<div class="card card-error">
					<h3>❌ Không tìm thấy thông tin lượt chạy</h3>
					<p>Lỗi: %v</p>
					<a href="#/dashboard" class="btn btn-primary" style="margin-top: 16px;">
						← Quay lại Bảng điều khiển
					</a>
				</div>
			`, errRun))
			return
		}

		veGiaoDienChiTietLuotChay(khungNoiDung, run)
	}()
}

/**
 * Render toàn bộ giao diện chi tiết 1 lần build và khung Terminal log.
 */
func veGiaoDienChiTietLuotChay(khungNoiDung dom.PhanTu, r *api.ThongTinLuotChay) {
	khungNoiDung.XoaHetCon()

	tenTrangThai, lopBadge, iconTrangThai := DinhDangTrangThai(r.Status)

	tenDuAn := "Dự án"
	duAnId := r.ProjectID
	linkDuAn := fmt.Sprintf("#/projects/%d", duAnId)
	if r.DuAn != nil {
		tenDuAn = r.DuAn.Name
	}

	hashDayDu := r.CommitHash
	hashNgan := r.CommitHash
	if len(hashNgan) > 7 {
		hashNgan = hashNgan[:7]
	}

	// 1. Breadcrumbs & Header
	theHeader := dom.TaoPhanTu("div").ThemLopCss("page-header")
	nutExternalHtml := ""
	if r.ExternalURL != "" {
		nutExternalHtml = fmt.Sprintf(`
			<a href="%s" target="_blank" rel="noopener noreferrer" class="btn btn-outline">
				<span>🌐</span> Xem trên CI Provider
			</a>
		`, r.ExternalURL)
	}

	theHeader.DatHtml(fmt.Sprintf(`
		<div class="header-left">
			<div class="breadcrumbs">
				<a href="#/dashboard" class="breadcrumb-item">Bảng điều khiển</a>
				<span class="breadcrumb-separator">/</span>
				<a href="%s" class="breadcrumb-item">%s</a>
				<span class="breadcrumb-separator">/</span>
				<span class="breadcrumb-item active">Build #%d</span>
			</div>
			<div class="project-title-row">
				<h1 class="page-title">%s #%d</h1>
				<span class="badge %s" style="font-size: 0.9rem; padding: 6px 14px;">%s %s</span>
			</div>
			<p class="page-subtitle">Thực thi tại %s • Thời lượng: %s</p>
		</div>
		<div class="header-actions">
			<a href="%s" class="btn btn-outline">
				← Quay lại Dự án
			</a>
			%s
		</div>
	`, linkDuAn, tenDuAn, r.ID,
		r.PipelineName, r.ID,
		lopBadge, iconTrangThai, tenTrangThai,
		DinhDangThoiGian(r.CreatedAt), DinhDangThoiLuong(r.DurationSeconds),
		linkDuAn, nutExternalHtml))
	khungNoiDung.ThemCon(theHeader)

	// 2. Thẻ Thông tin Chi tiết Thực thi (Execution Metadata Grid)
	theMetaGrid := dom.TaoPhanTu("div").ThemLopCss("card run-meta-grid")
	theMetaGrid.DatHtml(fmt.Sprintf(`
		<div class="run-meta-item">
			<span class="meta-label">Nhánh Git</span>
			<span class="meta-value"><strong>🌿 %s</strong></span>
		</div>
		<div class="run-meta-item">
			<span class="meta-label">Commit Hash</span>
			<span class="meta-value"><code>#%s</code> <small class="text-muted">(%s)</small></span>
		</div>
		<div class="run-meta-item" style="grid-column: span 2;">
			<span class="meta-label">Mô tả Commit</span>
			<span class="meta-value"><strong>💬 %s</strong></span>
		</div>
		<div class="run-meta-item">
			<span class="meta-label">Người kích hoạt</span>
			<span class="meta-value">👤 %s</span>
		</div>
		<div class="run-meta-item">
			<span class="meta-label">Sự kiện kích hoạt</span>
			<span class="meta-value">⚡ %s</span>
		</div>
		<div class="run-meta-item">
			<span class="meta-label">Thời điểm bắt đầu</span>
			<span class="meta-value text-muted">%s</span>
		</div>
		<div class="run-meta-item">
			<span class="meta-label">Thời điểm kết thúc</span>
			<span class="meta-value text-muted">%s</span>
		</div>
	`, r.Branch, hashNgan, hashDayDu,
		r.CommitMessage, r.Author, r.TriggerEvent,
		DinhDangThoiGian(r.StartedAt), DinhDangThoiGian(r.FinishedAt)))
	khungNoiDung.ThemCon(theMetaGrid)

	// 3. Trình Xem Log Terminal (Monospace Log Viewer)
	theTerminalSection := dom.TaoPhanTu("section").ThemLopCss("section-log-viewer")
	theTerminal := dom.TaoPhanTu("div").ThemLopCss("terminal-log")

	theTerminalHeader := dom.TaoPhanTu("div").ThemLopCss("terminal-header")
	theTerminalHeader.DatHtml(`
		<div class="terminal-title">
			<span class="terminal-dots">
				<span class="dot red"></span>
				<span class="dot yellow"></span>
				<span class="dot green"></span>
			</span>
			<span class="terminal-name">Pipeline Execution Log (Tối đa 64KB)</span>
		</div>
		<div class="terminal-actions">
			<span id="copy-log-status" class="text-muted" style="font-size: 0.8rem; margin-right: 8px;"></span>
			<button id="btn-copy-log" class="btn btn-outline btn-sm">📋 Sao chép Log</button>
		</div>
	`)
	theTerminal.ThemCon(theTerminalHeader)

	theTerminalBody := dom.TaoPhanTu("div").ThemLopCss("terminal-body").DatId("terminal-log-body")

	noiDungLog := strings.TrimSpace(r.ShortLog)
	if noiDungLog == "" {
		theTerminalBody.DatHtml(`
			<div class="terminal-empty text-muted">
				[DevOps Sentinel] Không có bản ghi log nào được lưu lại cho lượt chạy này.
			</div>
		`)
	} else {
		cacDong := strings.Split(noiDungLog, "\n")
		var b strings.Builder
		for stt, dong := range cacDong {
			lopDong := "log-line"
			dongLower := strings.ToLower(dong)
			if strings.Contains(dongLower, "error") ||
				strings.Contains(dongLower, "fail") ||
				strings.Contains(dongLower, "fatal") ||
				strings.Contains(dongLower, "exception") {
				lopDong += " log-line-error"
			} else if strings.Contains(dongLower, "success") ||
				strings.Contains(dongLower, "passed") ||
				strings.Contains(dongLower, "ok") {
				lopDong += " log-line-success"
			} else if strings.Contains(dongLower, "warn") {
				lopDong += " log-line-warning"
			}

			// Chống XSS cơ bản cho nội dung log
			dongAnToan := strings.ReplaceAll(dong, "<", "&lt;")
			dongAnToan = strings.ReplaceAll(dongAnToan, ">", "&gt;")

			b.WriteString(fmt.Sprintf(`<div class="%s"><span class="line-num">%d</span><span class="line-text">%s</span></div>`,
				lopDong, stt+1, dongAnToan))
		}
		theTerminalBody.DatHtml(b.String())
	}

	theTerminal.ThemCon(theTerminalBody)
	theTerminalSection.ThemCon(theTerminal)
	khungNoiDung.ThemCon(theTerminalSection)

	// Gắn sự kiện sao chép log
	nutCopyLog := theTerminalHeader.Tim("#btn-copy-log")
	if nutCopyLog.HopLe() {
		nutCopyLog.GanSuKien("click", func(this js.Value, args []js.Value) any {
			cuaSo := dom.LayCuaSo()
			if cuaSo.HopLe() {
				cuaSo.GiaTri.Get("navigator").Get("clipboard").Call("writeText", r.ShortLog)
				statusText := theTerminalHeader.Tim("#copy-log-status")
				if statusText.HopLe() {
					statusText.DatNoiDung("✅ Đã chép!")
				}
			}
			return nil
		})
	}
}
