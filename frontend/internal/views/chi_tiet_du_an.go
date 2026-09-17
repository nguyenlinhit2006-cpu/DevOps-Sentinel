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
 * Hiển thị màn hình Chi tiết Dự án (`#/projects/{id}`).
 * Bao gồm thông tin cấu hình webhook, ngưỡng cảnh báo, lịch sử pipeline runs có bộ lọc và phân trang.
 */
func HienThiTrangChiTietDuAn(idChuoi string) {
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
	khungNoiDung.DatId("project-detail-content")

	khungNoiDung.DatHtml(`
		<div class="content-loading">
			<div class="spinner"></div>
			<h3>Đang tải thông tin chi tiết dự án...</h3>
			<p>Vui lòng đợi giây lát trong khi WebAssembly tải dữ liệu.</p>
		</div>
	`)
	boKhung.ThemCon(khungNoiDung)
	khungChinh.ThemCon(boKhung)

	taiDuLieuChiTietDuAn(khungNoiDung, id, 1, "", "")
}

/**
 * Nạp dữ liệu chi tiết dự án và lịch sử lượt chạy từ API.
 */
func taiDuLieuChiTietDuAn(khungNoiDung dom.PhanTu, duAnId int64, trangHienTai int, nhanhLoc, trangThaiLoc string) {
	go func() {
		chiTiet, errChiTiet := api.LayChiTietDuAn(duAnId)
		danhSachRuns, meta, errRuns := api.LayLichSuLuotChayDuAn(duAnId, trangHienTai, 10, nhanhLoc, trangThaiLoc)

		if errChiTiet != nil {
			khungNoiDung.DatHtml(fmt.Sprintf(`
				<div class="card card-error">
					<h3>❌ Không tìm thấy thông tin dự án</h3>
					<p>Lỗi: %v</p>
					<a href="#/dashboard" class="btn btn-primary" style="margin-top: 16px;">
						← Quay lại Bảng điều khiển
					</a>
				</div>
			`, errChiTiet))
			return
		}

		veGiaoDienChiTietDuAn(khungNoiDung, chiTiet, danhSachRuns, meta, trangHienTai, nhanhLoc, trangThaiLoc, errRuns)
	}()
}

/**
 * Render toàn bộ giao diện chi tiết dự án.
 */
func veGiaoDienChiTietDuAn(khungNoiDung dom.PhanTu, da *api.ThongTinDuAnChiTiet, runs []api.ThongTinLuotChay, meta *api.PhanTrangMeta, trangHienTai int, nhanhLoc, trangThaiLoc string, errRuns error) {
	khungNoiDung.XoaHetCon()

	providerTen, providerLop, providerIcon := DinhDangNhaCungCap(da.CIProvider)

	// 1. Breadcrumbs & Header
	theHeader := dom.TaoPhanTu("div").ThemLopCss("page-header")
	theHeader.DatHtml(fmt.Sprintf(`
		<div class="header-left">
			<div class="breadcrumbs">
				<a href="#/dashboard" class="breadcrumb-item">Bảng điều khiển</a>
				<span class="breadcrumb-separator">/</span>
				<a href="#/projects" class="breadcrumb-item">Dự án</a>
				<span class="breadcrumb-separator">/</span>
				<span class="breadcrumb-item active">%s</span>
			</div>
			<div class="project-title-row">
				<h1 class="page-title">%s</h1>
				<span class="badge %s">%s %s</span>
			</div>
			<p class="page-subtitle">%s</p>
		</div>
		<div class="header-actions">
			<a href="%s" target="_blank" rel="noopener noreferrer" class="btn btn-outline">
				<span>🔗</span> Mở Kho Lưu Trữ
			</a>
		</div>
	`, da.Name, da.Name, providerLop, providerIcon, providerTen, da.Description, da.RepositoryURL))
	khungNoiDung.ThemCon(theHeader)

	// 2. Thẻ Thông tin Tổng quan & Thẻ Cấu hình Webhook
	khungTren := dom.TaoPhanTu("div").ThemLopCss("project-detail-grid")

	// Card Thông tin cơ bản
	tenNhom := "Chưa phân nhóm"
	if da.NhomSoHuu != nil {
		tenNhom = da.NhomSoHuu.Name
	}
	tenNguoiTao := "Hệ thống"
	if da.NguoiTao != nil {
		tenNguoiTao = da.NguoiTao.Name
	}

	theCanhBaoHtml := `<span class="badge badge-success-subtle">Không có lỗi</span>`
	if da.ActiveAlertsCount > 0 {
		theCanhBaoHtml = fmt.Sprintf(`<span class="badge badge-danger-pulse">🚨 %d Cảnh báo kích hoạt</span>`, da.ActiveAlertsCount)
	}

	cardThongTin := dom.TaoPhanTu("div").ThemLopCss("card info-card")
	cardThongTin.DatHtml(fmt.Sprintf(`
		<h3 class="card-title">📌 Thông Tin Dự Án</h3>
		<div class="meta-list">
			<div class="meta-item">
				<span class="meta-label">Đường dẫn Slug:</span>
				<span class="meta-value"><code>%s</code></span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Nhóm quản lý:</span>
				<span class="meta-value"><strong>%s</strong></span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Người khởi tạo:</span>
				<span class="meta-value">%s</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Ngưỡng cảnh báo:</span>
				<span class="meta-value">Fail liên tiếp <strong>%d lần</strong></span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Trạng thái Sentinel:</span>
				<span class="meta-value">%s</span>
			</div>
			<div class="meta-item">
				<span class="meta-label">Ngày khởi tạo:</span>
				<span class="meta-value text-muted">%s</span>
			</div>
		</div>
	`, da.Slug, tenNhom, tenNguoiTao, da.FailureThreshold, theCanhBaoHtml, DinhDangThoiGian(da.CreatedAt)))
	khungTren.ThemCon(cardThongTin)

	// Card Cấu hình Webhook
	cardWebhook := dom.TaoPhanTu("div").ThemLopCss("card webhook-card")
	cuaSo := dom.LayCuaSo()
	host := "localhost:8000"
	if cuaSo.HopLe() {
		host = cuaSo.GiaTri.Get("location").Get("hostname").String() + ":8000"
	}
	webhookUrl := fmt.Sprintf("http://%s/api/v1/webhooks/projects/%d", host, da.ID)

	nutSinhLaiHtml := ""
	if state.LayTrangThai().LaTruongNhom() {
		nutSinhLaiHtml = `<button id="btn-sinh-lai-secret" class="btn btn-outline btn-sm">🔄 Sinh lại Secret</button>`
	}

	cardWebhook.DatHtml(fmt.Sprintf(`
		<div class="webhook-header">
			<h3 class="card-title">⚡ Cấu Hình Webhook Ingestion</h3>
			%s
		</div>
		<p class="text-muted" style="font-size: 0.85rem; margin-bottom: 14px;">
			Cấu hình Webhook trên %s để tự động gửi dữ liệu build về Sentinel.
		</p>
		<div class="webhook-field">
			<label>Payload URL</label>
			<div class="input-copy-group">
				<input type="text" id="webhook-url-val" value="%s" readonly />
				<button id="btn-copy-webhook-url" class="btn btn-outline btn-sm" title="Sao chép URL">📋 Copy</button>
			</div>
		</div>
		<div class="webhook-field">
			<label>Webhook Secret (HMAC SHA-256)</label>
			<div class="input-copy-group">
				<input type="password" id="webhook-secret-val" value="%s" readonly />
				<button id="btn-toggle-secret" class="btn btn-outline btn-sm" title="Ẩn/Hiện Secret">👁️ Hiện</button>
				<button id="btn-copy-secret" class="btn btn-outline btn-sm" title="Sao chép Secret">📋 Copy</button>
			</div>
		</div>
		<div id="webhook-alert-msg" class="alert-box" style="display: none; margin-top: 10px;"></div>
	`, nutSinhLaiHtml, providerTen, webhookUrl, da.WebhookSecret))
	khungTren.ThemCon(cardWebhook)

	khungNoiDung.ThemCon(khungTren)

	// Gắn sự kiện sao chép Webhook URL
	saoChepVanBan := func(vanBan, thongDiep string) {
		cuaSo := dom.LayCuaSo()
		if cuaSo.HopLe() {
			cuaSo.GiaTri.Get("navigator").Get("clipboard").Call("writeText", vanBan)
			theMsg := cardWebhook.Tim("#webhook-alert-msg")
			if theMsg.HopLe() {
				theMsg.DatThuocTinh("style", "display: block;").ThemLopCss("alert-success")
				theMsg.DatNoiDung(thongDiep)
			}
		}
	}

	nutCopyUrl := cardWebhook.Tim("#btn-copy-webhook-url")
	if nutCopyUrl.HopLe() {
		nutCopyUrl.GanSuKien("click", func(this js.Value, args []js.Value) any {
			saoChepVanBan(webhookUrl, "✅ Đã sao chép Webhook URL vào bộ nhớ tạm!")
			return nil
		})
	}

	nutCopySecret := cardWebhook.Tim("#btn-copy-secret")
	if nutCopySecret.HopLe() {
		nutCopySecret.GanSuKien("click", func(this js.Value, args []js.Value) any {
			saoChepVanBan(da.WebhookSecret, "✅ Đã sao chép Webhook Secret vào bộ nhớ tạm!")
			return nil
		})
	}

	nutToggleSecret := cardWebhook.Tim("#btn-toggle-secret")
	if nutToggleSecret.HopLe() {
		nutToggleSecret.GanSuKien("click", func(this js.Value, args []js.Value) any {
			oSecret := cardWebhook.Tim("#webhook-secret-val")
			if oSecret.HopLe() {
				loaiHienTai := oSecret.GiaTri.Get("type").String()
				if loaiHienTai == "password" {
					oSecret.DatThuocTinh("type", "text")
					nutToggleSecret.DatHtml("🔒 Ẩn")
				} else {
					oSecret.DatThuocTinh("type", "password")
					nutToggleSecret.DatHtml("👁️ Hiện")
				}
			}
			return nil
		})
	}

	nutSinhLaiSecret := cardWebhook.Tim("#btn-sinh-lai-secret")
	if nutSinhLaiSecret.HopLe() {
		nutSinhLaiSecret.GanSuKien("click", func(this js.Value, args []js.Value) any {
			nutSinhLaiSecret.DatThuocTinh("disabled", "true").DatHtml(`<span class="spinner-sm"></span> Đang sinh lại...`)
			go func() {
				secretMoi, err := api.TaoLaiWebhookSecret(da.ID)
				nutSinhLaiSecret.XoaThuocTinh("disabled").DatHtml("🔄 Sinh lại Secret")
				theMsg := cardWebhook.Tim("#webhook-alert-msg")
				if err != nil {
					if theMsg.HopLe() {
						theMsg.DatThuocTinh("style", "display: block;").ThemLopCss("alert-error")
						theMsg.DatNoiDung(fmt.Sprintf("❌ Lỗi: %v", err))
					}
				} else {
					da.WebhookSecret = secretMoi
					oSecret := cardWebhook.Tim("#webhook-secret-val")
					if oSecret.HopLe() {
						oSecret.DatGiaTri(secretMoi)
					}
					if theMsg.HopLe() {
						theMsg.DatThuocTinh("style", "display: block;").ThemLopCss("alert-success")
						theMsg.DatNoiDung("✅ Đã sinh lại Webhook Secret thành công! Hãy cập nhật CI provider của bạn.")
					}
				}
			}()
			return nil
		})
	}

	// 3. Phân vùng Bảng Lịch sử Lượt chạy Pipeline
	phanVungHistory := dom.TaoPhanTu("section").ThemLopCss("section-build-history")
	veBangLichSuBuild(phanVungHistory, khungNoiDung, da.ID, runs, meta, trangHienTai, nhanhLoc, trangThaiLoc, errRuns)
	khungNoiDung.ThemCon(phanVungHistory)
}

/**
 * Vẽ bảng danh sách lịch sử lượt chạy pipeline kèm bộ lọc và thanh phân trang.
 */
func veBangLichSuBuild(phanVung dom.PhanTu, khungNoiDung dom.PhanTu, duAnId int64, runs []api.ThongTinLuotChay, meta *api.PhanTrangMeta, trangHienTai int, nhanhLoc, trangThaiLoc string, errRuns error) {
	headerHtml := `
		<div class="section-header">
			<div class="section-title-wrap">
				<h2 class="section-title">Lịch Sử Lượt Chạy Pipeline</h2>
				<span class="badge badge-count">Pipeline Runs</span>
			</div>
		</div>
	`
	phanVung.DatHtml(headerHtml)

	// Thanh lọc theo nhánh và trạng thái
	theFilter := dom.TaoPhanTu("div").ThemLopCss("card filter-toolbar")
	theFilter.DatHtml(fmt.Sprintf(`
		<div class="filter-controls">
			<div class="filter-group">
				<label>Trạng thái:</label>
				<select id="select-loc-status" class="filter-select">
					<option value="all" %s>Tất cả trạng thái</option>
					<option value="success" %s>✓ Thành công</option>
					<option value="failed" %s>✗ Thất bại</option>
					<option value="running" %s>⚡ Đang chạy</option>
					<option value="queued" %s>⏳ Đang đợi</option>
				</select>
			</div>
			<div class="filter-group">
				<label>Nhánh:</label>
				<input type="text" id="input-loc-nhanh" class="filter-input" placeholder="main, develop..." value="%s" />
			</div>
			<button id="btn-ap-dung-loc" class="btn btn-primary btn-sm">Áp dụng</button>
		</div>
	`, chonTuyChon(trangThaiLoc, "all"), chonTuyChon(trangThaiLoc, "success"),
		chonTuyChon(trangThaiLoc, "failed"), chonTuyChon(trangThaiLoc, "running"),
		chonTuyChon(trangThaiLoc, "queued"), nhanhLoc))
	phanVung.ThemCon(theFilter)

	// Gắn sự kiện nút lọc
	nutLoc := theFilter.Tim("#btn-ap-dung-loc")
	if nutLoc.HopLe() {
		nutLoc.GanSuKien("click", func(this js.Value, args []js.Value) any {
			trangThaiChon := theFilter.Tim("#select-loc-status").LayGiaTri()
			nhanhChon := strings.TrimSpace(theFilter.Tim("#input-loc-nhanh").LayGiaTri())
			taiDuLieuChiTietDuAn(khungNoiDung, duAnId, 1, nhanhChon, trangThaiChon)
			return nil
		})
	}

	// Bảng dữ liệu
	theCardTable := dom.TaoPhanTu("div").ThemLopCss("card table-card")

	if errRuns != nil {
		theCardTable.DatHtml(fmt.Sprintf(`<div class="alert-box alert-error">Lỗi tải lịch sử build: %v</div>`, errRuns))
		phanVung.ThemCon(theCardTable)
		return
	}

	if len(runs) == 0 {
		theCardTable.DatHtml(`
			<div class="empty-state">
				<div class="empty-icon">⏳</div>
				<h3>Chưa có lượt chạy nào</h3>
				<p class="text-muted">Chưa ghi nhận pipeline run nào khớp với điều kiện lọc.</p>
			</div>
		`)
		phanVung.ThemCon(theCardTable)
		return
	}

	theTable := dom.TaoPhanTu("div").ThemLopCss("table-responsive")
	theTable.DatHtml(`
		<table class="runs-table">
			<thead>
				<tr>
					<th>Trạng Thái</th>
					<th>Pipeline & Sự Kiện</th>
					<th>Nhánh & Commit</th>
					<th>Mô Tả Commit</th>
					<th>Người Chạy</th>
					<th>Thời Lượng</th>
					<th>Thời Gian</th>
					<th>Thao Tác</th>
				</tr>
			</thead>
			<tbody id="tbody-runs"></tbody>
		</table>
	`)
	theCardTable.ThemCon(theTable)

	tbody := theTable.Tim("#tbody-runs")
	if tbody.HopLe() {
		for _, r := range runs {
			hang := dom.TaoPhanTu("tr")
			tenTrangThai, lopBadge, iconTrangThai := DinhDangTrangThai(r.Status)
			hashNgan := r.CommitHash
			if len(hashNgan) > 7 {
				hashNgan = hashNgan[:7]
			}
			msg := r.CommitMessage
			if len(msg) > 40 {
				msg = msg[:38] + "..."
			}

			hang.DatHtml(fmt.Sprintf(`
				<td>
					<span class="badge %s">%s %s</span>
				</td>
				<td>
					<strong>%s</strong>
					<div class="text-muted" style="font-size: 0.78rem;">Sự kiện: %s</div>
				</td>
				<td>
					<span class="run-branch">🌿 %s</span>
					<div><code>#%s</code></div>
				</td>
				<td class="text-muted" title="%s">%s</td>
				<td>👤 %s</td>
				<td>⏱️ %s</td>
				<td class="text-muted" style="font-size: 0.8rem;">%s</td>
				<td>
					<a href="#/runs/%d" class="btn btn-outline btn-sm">
						Xem log →
					</a>
				</td>
			`, lopBadge, iconTrangThai, tenTrangThai,
				r.PipelineName, r.TriggerEvent,
				r.Branch, hashNgan,
				r.CommitMessage, msg,
				r.Author, DinhDangThoiLuong(r.DurationSeconds),
				DinhDangThoiGian(r.CreatedAt),
				r.ID))

			tbody.ThemCon(hang)
		}
	}

	// 4. Thanh Phân Trang (Pagination Controls)
	if meta != nil && meta.TotalPages > 1 {
		thePhanTrang := dom.TaoPhanTu("div").ThemLopCss("pagination-bar")

		lopNutTruoc := ""
		if trangHienTai <= 1 {
			lopNutTruoc = "disabled"
		}
		lopNutSau := ""
		if trangHienTai >= meta.TotalPages {
			lopNutSau = "disabled"
		}

		thePhanTrang.DatHtml(fmt.Sprintf(`
			<button id="btn-trang-truoc" class="btn btn-outline btn-sm %s">← Trang trước</button>
			<span class="page-indicator">Trang <strong>%d</strong> / %d (Tổng %d lượt build)</span>
			<button id="btn-trang-sau" class="btn btn-outline btn-sm %s">Trang sau →</button>
		`, lopNutTruoc, trangHienTai, meta.TotalPages, meta.Total, lopNutSau))

		theCardTable.ThemCon(thePhanTrang)

		nutTruoc := thePhanTrang.Tim("#btn-trang-truoc")
		if nutTruoc.HopLe() && trangHienTai > 1 {
			nutTruoc.GanSuKien("click", func(this js.Value, args []js.Value) any {
				taiDuLieuChiTietDuAn(khungNoiDung, duAnId, trangHienTai-1, nhanhLoc, trangThaiLoc)
				return nil
			})
		}

		nutSau := thePhanTrang.Tim("#btn-trang-sau")
		if nutSau.HopLe() && trangHienTai < meta.TotalPages {
			nutSau.GanSuKien("click", func(this js.Value, args []js.Value) any {
				taiDuLieuChiTietDuAn(khungNoiDung, duAnId, trangHienTai+1, nhanhLoc, trangThaiLoc)
				return nil
			})
		}
	}

	phanVung.ThemCon(theCardTable)
}

func chonTuyChon(giaTri, mucTieu string) string {
	if giaTri == mucTieu {
		return "selected"
	}
	return ""
}
