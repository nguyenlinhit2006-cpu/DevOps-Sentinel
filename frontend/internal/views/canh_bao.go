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

// Biến lưu trạng thái tab bộ lọc cảnh báo đang chọn (active hoặc resolved)
var tabCanhBaoHienTai = "active"

/**
 * Hiển thị màn hình Quản lý Cảnh báo Sentinel (`#/alerts`).
 * Cho phép xem các cảnh báo đang mở (Active) do pipeline thất bại liên tiếp,
 * xem lịch sử cảnh báo đã đóng (Resolved), và thao tác đóng cảnh báo (Team Lead/Admin).
 */
func HienThiTrangCanhBao() {
	if !state.LayTrangThai().DaDangNhap() {
		routing.ChuyenHuong("#/login")
		return
	}

	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	boKhung := dom.TaoPhanTu("div").ThemLopCss("dashboard-layout")
	thanhNav := TaoThanhDieuHuong("alerts")
	boKhung.ThemCon(thanhNav)

	khungNoiDung := dom.TaoPhanTu("main").ThemLopCss("app-content")
	khungNoiDung.DatId("alerts-content")

	khungNoiDung.DatHtml(`
		<div class="content-header">
			<div class="header-titles">
				<h1 class="page-title">🚨 Cảnh báo Sentinel</h1>
				<p class="page-subtitle">Giám sát các sự cố pipeline CI/CD vi phạm ngưỡng thất bại liên tiếp và theo dõi lịch sử xử lý.</p>
			</div>
			<div class="header-actions">
				<button id="btn-lam-moi-canh-bao" class="btn btn-outline">
					<span>🔄</span> Làm mới
				</button>
			</div>
		</div>

		<!-- Thanh Tabs lọc trạng thái cảnh báo -->
		<div class="alerts-filter-bar">
			<div class="tabs-segmented">
				<button id="tab-alerts-active" class="tab-btn active">
					<span class="tab-indicator-dot danger"></span> Cảnh Báo Đang Mở (Active)
				</button>
				<button id="tab-alerts-resolved" class="tab-btn">
					<span class="tab-indicator-dot success"></span> Lịch Sử Đã Đóng (Resolved)
				</button>
			</div>
			<div id="alerts-count-badge" class="alerts-summary-meta">
				Đang tải số liệu...
			</div>
		</div>

		<!-- Vùng hiển thị danh sách thẻ cảnh báo -->
		<div id="alerts-list-container" class="alerts-container">
			<div class="content-loading">
				<div class="spinner"></div>
				<p>Đang tải danh sách cảnh báo...</p>
			</div>
		</div>
	`)

	boKhung.ThemCon(khungNoiDung)
	khungChinh.ThemCon(boKhung)

	// Gán sự kiện cho nút Làm mới
	nutLamMoi := khungNoiDung.Tim("#btn-lam-moi-canh-bao")
	if nutLamMoi.HopLe() {
		nutLamMoi.GanSuKien("click", func(this js.Value, args []js.Value) any {
			taiDuLieuCanhBao(khungNoiDung, tabCanhBaoHienTai)
			return nil
		})
	}

	// Gán sự kiện chuyển Tab Active
	tabActive := khungNoiDung.Tim("#tab-alerts-active")
	tabResolved := khungNoiDung.Tim("#tab-alerts-resolved")

	if tabActive.HopLe() && tabResolved.HopLe() {
		tabActive.GanSuKien("click", func(this js.Value, args []js.Value) any {
			if tabCanhBaoHienTai != "active" {
				tabCanhBaoHienTai = "active"
				tabActive.ThemLopCss("active")
				tabResolved.XoaLopCss("active")
				taiDuLieuCanhBao(khungNoiDung, "active")
			}
			return nil
		})

		tabResolved.GanSuKien("click", func(this js.Value, args []js.Value) any {
			if tabCanhBaoHienTai != "resolved" {
				tabCanhBaoHienTai = "resolved"
				tabResolved.ThemLopCss("active")
				tabActive.XoaLopCss("active")
				taiDuLieuCanhBao(khungNoiDung, "resolved")
			}
			return nil
		})
	}

	// Tải dữ liệu lần đầu
	taiDuLieuCanhBao(khungNoiDung, tabCanhBaoHienTai)
}

/**
 * Tải danh sách cảnh báo từ API theo trạng thái được chọn (active hoặc resolved).
 */
func taiDuLieuCanhBao(khungNoiDung dom.PhanTu, trangThai string) {
	khungDanhSach := khungNoiDung.Tim("#alerts-list-container")
	theDem := khungNoiDung.Tim("#alerts-count-badge")

	if !khungDanhSach.HopLe() {
		return
	}

	go func() {
		danhSach, err := api.LayDanhSachCanhBao(trangThai, 0)
		if err != nil {
			khungDanhSach.DatHtml(fmt.Sprintf(`
				<div class="card card-error">
					<h3>⚠️ Không thể tải danh sách cảnh báo</h3>
					<p>%s</p>
					<button class="btn btn-outline" onclick="location.reload()" style="margin-top: 12px;">Thử lại</button>
				</div>
			`, err.Error()))
			return
		}

		soLuong := len(danhSach)
		if theDem.HopLe() {
			if trangThai == "active" {
				if soLuong > 0 {
					theDem.DatHtml(fmt.Sprintf(`<span class="badge badge-failed">%d cảnh báo đang kích hoạt</span>`, soLuong))
				} else {
					theDem.DatHtml(`<span class="badge badge-success">Hệ thống an toàn (0)</span>`)
				}
			} else {
				theDem.DatHtml(fmt.Sprintf(`<span class="badge badge-neutral">%d cảnh báo đã xử lý</span>`, soLuong))
			}
		}

		if soLuong == 0 {
			if trangThai == "active" {
				khungDanhSach.DatHtml(`
					<div class="empty-state-card card">
						<div class="empty-icon">🛡️</div>
						<h3>Không có cảnh báo sự cố nào đang mở!</h3>
						<p>Tất cả các pipeline CI/CD đều đang vận hành trong ngưỡng an toàn hoặc chưa gặp chuỗi lỗi liên tiếp.</p>
					</div>
				`)
			} else {
				khungDanhSach.DatHtml(`
					<div class="empty-state-card card">
						<div class="empty-icon">📋</div>
						<h3>Chưa có cảnh báo nào trong lịch sử</h3>
						<p>Khi các cảnh báo sự cố được khắc phục và đóng lại, thông tin lưu trữ sẽ hiển thị tại đây.</p>
					</div>
				`)
			}
			return
		}

		// Xây dựng danh sách thẻ cảnh báo
		var sb strings.Builder
		sb.WriteString(`<div class="alerts-grid">`)

		coQuyenDong := state.LayTrangThai().LaTruongNhom()

		for _, cb := range danhSach {
			tenDuAn := "Dự án #" + fmt.Sprintf("%d", cb.ProjectID)
			if cb.DuAn != nil && cb.DuAn.Name != "" {
				tenDuAn = cb.DuAn.Name
			}

			// Chi tiết commit & lượt chạy lỗi
			thongTinLoi := "Không rõ chi tiết lượt chạy"
			linkLuotChay := ""
			if cb.LuotChayLoi != nil {
				hashRutGon := cb.LuotChayLoi.CommitHash
				if len(hashRutGon) > 7 {
					hashRutGon = hashRutGon[:7]
				}
				thongTinLoi = fmt.Sprintf("Nhánh: <code>%s</code> | Commit: <code>%s</code>", cb.LuotChayLoi.Branch, hashRutGon)
				linkLuotChay = fmt.Sprintf(`<a href="#/runs/%d" class="btn btn-outline btn-xs" title="Xem chi tiết log console">🔍 Xem Run #%d</a>`, cb.LuotChayLoi.ID, cb.LuotChayLoi.ID)
			}

			if cb.Status == "active" {
				// Thẻ Cảnh báo Đang Kích Hoạt (Hazard Pulse Card)
				nutDong := ""
				if coQuyenDong {
					nutDong = fmt.Sprintf(`
						<button class="btn btn-success btn-sm btn-dong-canh-bao" data-alert-id="%d">
							<span>✓</span> Xác nhận Đóng cảnh báo
						</button>
					`, cb.ID)
				}

				sb.WriteString(fmt.Sprintf(`
					<div class="card alert-hazard-card pulse-danger" id="alert-card-%d">
						<div class="alert-card-header">
							<div class="alert-status-badge">
								<span class="hazard-dot"></span>
								<span class="hazard-title">NGUY HIỂM: PIPELINE THẤT BẠI LIÊN TIẾP</span>
							</div>
							<span class="alert-failures-count">🔥 %d Lần Thất Bại</span>
						</div>

						<div class="alert-card-body">
							<h3 class="alert-project-title">
								<a href="#/projects/%d">📁 %s</a>
							</h3>
							<h4 class="alert-headline">%s</h4>
							<p class="alert-message">%s</p>

							<div class="alert-meta-box">
								<div class="meta-item">
									<span class="meta-label">Sự cố gần nhất:</span>
									<span class="meta-val">%s</span>
								</div>
								<div class="meta-item">
									<span class="meta-label">Thời điểm phát hiện:</span>
									<span class="meta-val">%s</span>
								</div>
							</div>
						</div>

						<div class="alert-card-footer">
							<div class="footer-actions-left">
								%s
							</div>
							<div class="footer-actions-right">
								%s
							</div>
						</div>
					</div>
				`, cb.ID, cb.ConsecutiveFailures, cb.ProjectID, tenDuAn, cb.Title, cb.Message,
					thongTinLoi, cb.CreatedAt, linkLuotChay, nutDong))

			} else {
				// Thẻ Cảnh báo Đã Khắc Phục (Resolved)
				nguoiXuLy := "Hệ thống"
				if cb.NguoiXuLy != nil && cb.NguoiXuLy.Name != "" {
					nguoiXuLy = cb.NguoiXuLy.Name
				}
				thoiGianDong := "Chưa rõ"
				if cb.ResolvedAt != nil {
					thoiGianDong = *cb.ResolvedAt
				}

				sb.WriteString(fmt.Sprintf(`
					<div class="card alert-resolved-card">
						<div class="alert-card-header">
							<div class="alert-status-badge resolved">
								<span>✓</span>
								<span class="resolved-title">ĐÃ KHẮC PHỤC SỰ CỐ</span>
							</div>
							<span class="badge badge-success">Resolved</span>
						</div>

						<div class="alert-card-body">
							<h3 class="alert-project-title">
								<a href="#/projects/%d">📁 %s</a>
							</h3>
							<h4 class="alert-headline">%s</h4>
							<p class="alert-message-dimmed">%s</p>

							<div class="alert-meta-box resolved-box">
								<div class="meta-item">
									<span class="meta-label">Người xử lý đóng cảnh báo:</span>
									<span class="meta-val text-bold">%s</span>
								</div>
								<div class="meta-item">
									<span class="meta-label">Thời điểm đóng:</span>
									<span class="meta-val">%s</span>
								</div>
							</div>
						</div>

						<div class="alert-card-footer">
							<div class="footer-actions-left">
								%s
							</div>
							<div class="footer-actions-right">
								<span class="text-muted text-xs">Đã ghi nhận vào Audit Log</span>
							</div>
						</div>
					</div>
				`, cb.ProjectID, tenDuAn, cb.Title, cb.Message, nguoiXuLy, thoiGianDong, linkLuotChay))
			}
		}

		sb.WriteString(`</div>`)
		khungDanhSach.DatHtml(sb.String())

		// Gán sự kiện click cho các nút Xác nhận Đóng cảnh báo
		for _, cb := range danhSach {
			if cb.Status == "active" {
				idCanhBao := cb.ID
				nut := khungDanhSach.Tim(fmt.Sprintf(`button[data-alert-id="%d"]`, idCanhBao))
				if nut.HopLe() {
					nut.GanSuKien("click", func(this js.Value, args []js.Value) any {
						xacNhan := js.Global().Call("confirm", fmt.Sprintf("Bạn có chắc chắn muốn đóng và xác nhận đã khắc phục Cảnh báo #%d?", idCanhBao))
						if !xacNhan.Bool() {
							return nil
						}

						nut.DatThuocTinh("disabled", "true")
						nut.DatHtml("Đang đóng...")

						go func() {
							err := api.DongCanhBao(idCanhBao)
							if err != nil {
								js.Global().Call("alert", "Lỗi: "+err.Error())
								nut.XoaThuocTinh("disabled")
								nut.DatHtml("<span>✓</span> Xác nhận Đóng cảnh báo")
								return
							}

							// Thông báo thành công và nạp lại danh sách cảnh báo
							taiDuLieuCanhBao(khungNoiDung, "active")
						}()

						return nil
					})
				}
			}
		}
	}()
}

/**
 * Làm mới dữ liệu cảnh báo một cách êm dịu (dành cho bộ realtime polling 4s ngầm).
 */
func LamMoiCanhBaoYenLang() {
	khungNoiDung := dom.LayPhanTuTheoId("alerts-content")
	if !khungNoiDung.HopLe() {
		return
	}

	taiDuLieuCanhBao(khungNoiDung, tabCanhBaoHienTai)
}

