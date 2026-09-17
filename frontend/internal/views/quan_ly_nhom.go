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
 * Hiển thị màn hình Danh sách Nhóm làm việc (`#/teams`).
 * Hiển thị danh sách các nhóm làm việc, số lượng thành viên, số lượng dự án trực thuộc
 * và cho phép Quản trị viên (Admin) tạo nhóm mới hoặc xóa nhóm.
 */
func HienThiTrangDanhSachNhom() {
	if !state.LayTrangThai().DaDangNhap() {
		routing.ChuyenHuong("#/login")
		return
	}

	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	boKhung := dom.TaoPhanTu("div").ThemLopCss("dashboard-layout")
	thanhNav := TaoThanhDieuHuong("teams")
	boKhung.ThemCon(thanhNav)

	khungNoiDung := dom.TaoPhanTu("main").ThemLopCss("app-content")
	khungNoiDung.DatId("teams-content")

	laAdmin := state.LayTrangThai().LaQuanTriVien()
	nutTaoNhomHtml := ""
	if laAdmin {
		nutTaoNhomHtml = `
			<button id="btn-mo-modal-tao-nhom" class="btn btn-primary">
				<span>➕</span> Tạo Nhóm Mới
			</button>
		`
	}

	khungNoiDung.DatHtml(fmt.Sprintf(`
		<div class="content-header">
			<div class="header-titles">
				<h1 class="page-title">👥 Quản lý Nhóm làm việc</h1>
				<p class="page-subtitle">Tổ chức phân bổ nhân sự, quản trị dự án theo nhóm và phân chia vai trò vận hành pipeline.</p>
			</div>
			<div class="header-actions">
				%s
				<button id="btn-lam-moi-nhom" class="btn btn-outline">
					<span>🔄</span> Làm mới
				</button>
			</div>
		</div>

		<!-- Khung Form tạo nhóm mới (Ẩn mặc định) -->
		<div id="modal-tao-nhom-khung" class="card modal-inline-card" style="display: none; margin-bottom: 24px;">
			<div class="card-header-flex">
				<h3 style="font-size: 1.15rem; color: #f8fafc;">✨ Tạo Nhóm Làm Việc Mới</h3>
				<button id="btn-dong-modal-tao-nhom" class="btn btn-outline btn-xs">✖ Đóng</button>
			</div>
			<div style="margin-top: 16px;">
				<div class="form-group">
					<label class="form-label" for="inp-ten-nhom">Tên nhóm làm việc <span class="required">*</span></label>
					<input id="inp-ten-nhom" type="text" class="form-control" placeholder="Ví dụ: Đội Ngũ Hạ Tầng & SRE" maxlength="100" />
				</div>
				<div class="form-group" style="margin-top: 12px;">
					<label class="form-label" for="inp-mo-ta-nhom">Mô tả nhiệm vụ của nhóm</label>
					<textarea id="inp-mo-ta-nhom" class="form-control" rows="2" placeholder="Ví dụ: Phụ trách hạ tầng đám mây, Kubernetes cluster và pipeline CI/CD core."></textarea>
				</div>
				<div id="thong-bao-loi-tao-nhom" class="alert-error-box" style="display: none; margin-top: 12px;"></div>
				<div class="form-actions" style="margin-top: 16px; display: flex; gap: 10px;">
					<button id="btn-xac-nhan-tao-nhom" class="btn btn-primary">
						<span>💾</span> Xác nhận Tạo Nhóm
					</button>
					<button id="btn-huy-tao-nhom" class="btn btn-outline">Hủy</button>
				</div>
			</div>
		</div>

		<!-- Vùng lưới danh sách các nhóm -->
		<div id="teams-grid-container" class="teams-container">
			<div class="content-loading">
				<div class="spinner"></div>
				<p>Đang nạp danh sách nhóm làm việc...</p>
			</div>
		</div>
	`, nutTaoNhomHtml))

	boKhung.ThemCon(khungNoiDung)
	khungChinh.ThemCon(boKhung)

	// Xử lý đóng/mở form tạo nhóm mới
	khungModal := khungNoiDung.Tim("#modal-tao-nhom-khung")
	btnMoModal := khungNoiDung.Tim("#btn-mo-modal-tao-nhom")
	btnDongModal := khungNoiDung.Tim("#btn-dong-modal-tao-nhom")
	btnHuyModal := khungNoiDung.Tim("#btn-huy-tao-nhom")

	if btnMoModal.HopLe() && khungModal.HopLe() {
		btnMoModal.GanSuKien("click", func(this js.Value, args []js.Value) any {
			khungModal.XoaLopCss("an-form")
			khungModal.DatThuocTinh("style", "display: block; margin-bottom: 24px;")
			inpTen := khungNoiDung.Tim("#inp-ten-nhom")
			if inpTen.HopLe() {
				inpTen.GiaTri.Call("focus")
			}
			return nil
		})
	}

	dongModal := func() {
		if khungModal.HopLe() {
			khungModal.DatThuocTinh("style", "display: none; margin-bottom: 24px;")
		}
	}

	if btnDongModal.HopLe() {
		btnDongModal.GanSuKien("click", func(this js.Value, args []js.Value) any {
			dongModal()
			return nil
		})
	}
	if btnHuyModal.HopLe() {
		btnHuyModal.GanSuKien("click", func(this js.Value, args []js.Value) any {
			dongModal()
			return nil
		})
	}

	// Xử lý tạo nhóm mới
	btnXacNhan := khungNoiDung.Tim("#btn-xac-nhan-tao-nhom")
	if btnXacNhan.HopLe() {
		btnXacNhan.GanSuKien("click", func(this js.Value, args []js.Value) any {
			inpTen := khungNoiDung.Tim("#inp-ten-nhom")
			inpMoTa := khungNoiDung.Tim("#inp-mo-ta-nhom")
			hopLoi := khungNoiDung.Tim("#thong-bao-loi-tao-nhom")

			tenNhom := strings.TrimSpace(inpTen.LayGiaTri())
			moTaNhom := strings.TrimSpace(inpMoTa.LayGiaTri())

			if tenNhom == "" {
				if hopLoi.HopLe() {
					hopLoi.DatThuocTinh("style", "display: block; margin-top: 12px;")
					hopLoi.DatHtml("Vui lòng nhập tên nhóm làm việc.")
				}
				return nil
			}

			if hopLoi.HopLe() {
				hopLoi.DatThuocTinh("style", "display: none;")
			}

			btnXacNhan.DatThuocTinh("disabled", "true")
			btnXacNhan.DatHtml("Đang tạo...")

			go func() {
				_, err := api.TaoNhomMoi(tenNhom, moTaNhom)
				btnXacNhan.XoaThuocTinh("disabled")
				btnXacNhan.DatHtml("<span>💾</span> Xác nhận Tạo Nhóm")

				if err != nil {
					if hopLoi.HopLe() {
						hopLoi.DatThuocTinh("style", "display: block; margin-top: 12px;")
						hopLoi.DatHtml("Lỗi: " + err.Error())
					}
					return
				}

				// Reset form và đóng modal
				inpTen.DatGiaTri("")
				inpMoTa.DatGiaTri("")
				dongModal()

				// Tải lại danh sách nhóm
				taiDanhSachNhom(khungNoiDung)
			}()

			return nil
		})
	}

	// Gán nút làm mới
	btnLamMoi := khungNoiDung.Tim("#btn-lam-moi-nhom")
	if btnLamMoi.HopLe() {
		btnLamMoi.GanSuKien("click", func(this js.Value, args []js.Value) any {
			taiDanhSachNhom(khungNoiDung)
			return nil
		})
	}

	// Tải danh sách nhóm
	taiDanhSachNhom(khungNoiDung)
}

/**
 * Tải và hiển thị danh sách nhóm từ API.
 */
func taiDanhSachNhom(khungNoiDung dom.PhanTu) {
	khungGrid := khungNoiDung.Tim("#teams-grid-container")
	if !khungGrid.HopLe() {
		return
	}

	go func() {
		danhSach, err := api.LayDanhSachNhom()
		if err != nil {
			khungGrid.DatHtml(fmt.Sprintf(`
				<div class="card card-error">
					<h3>⚠️ Không thể tải danh sách nhóm</h3>
					<p>%s</p>
					<button class="btn btn-outline" onclick="location.reload()" style="margin-top: 12px;">Thử lại</button>
				</div>
			`, err.Error()))
			return
		}

		if len(danhSach) == 0 {
			khungGrid.DatHtml(`
				<div class="empty-state-card card">
					<div class="empty-icon">👥</div>
					<h3>Chưa có nhóm làm việc nào</h3>
					<p>Hãy tạo nhóm đầu tiên để phân bổ thành viên và dự án vào quản lý tập trung.</p>
				</div>
			`)
			return
		}

		laAdmin := state.LayTrangThai().LaQuanTriVien()
		var sb strings.Builder
		sb.WriteString(`<div class="teams-grid">`)

		for _, nhom := range danhSach {
			moTa := nhom.Description
			if moTa == "" {
				moTa = "Chưa có mô tả cho nhóm này."
			}

			nguoiTao := "Quản trị viên"
			if nhom.NguoiTao != nil && nhom.NguoiTao.Name != "" {
				nguoiTao = nhom.NguoiTao.Name
			}

			nutXoa := ""
			if laAdmin {
				nutXoa = fmt.Sprintf(`
					<button class="btn btn-outline btn-xs btn-xoa-nhom" data-team-id="%d" data-team-name="%s" title="Xóa nhóm này">
						🗑️
					</button>
				`, nhom.ID, nhom.Name)
			}

			sb.WriteString(fmt.Sprintf(`
				<div class="card team-card" id="team-card-%d">
					<div class="team-card-header">
						<div class="team-avatar-badge">
							<span>👥</span>
						</div>
						<div class="team-header-main">
							<h3 class="team-card-title">
								<a href="#/teams/%d">%s</a>
							</h3>
							<span class="team-meta-author">Người tạo: %s</span>
						</div>
						<div class="team-header-actions">
							%s
						</div>
					</div>

					<p class="team-card-desc">%s</p>

					<div class="team-card-stats">
						<div class="team-stat-item">
							<span class="team-stat-num">%d</span>
							<span class="team-stat-label">Thành viên</span>
						</div>
						<div class="team-stat-divider"></div>
						<div class="team-stat-item">
							<span class="team-stat-num">%d</span>
							<span class="team-stat-label">Dự án CI/CD</span>
						</div>
					</div>

					<div class="team-card-footer">
						<a href="#/teams/%d" class="btn btn-outline btn-block">
							Xem Chi Tiết & Thành Viên →
						</a>
					</div>
				</div>
			`, nhom.ID, nhom.ID, nhom.Name, nguoiTao, nutXoa, moTa, nhom.MembersCount, nhom.ProjectsCount, nhom.ID))
		}

		sb.WriteString(`</div>`)
		khungGrid.DatHtml(sb.String())

		// Gán sự kiện cho nút xóa nhóm (nếu là admin)
		if laAdmin {
			for _, nhom := range danhSach {
				idNhom := nhom.ID
				tenNhom := nhom.Name
				nut := khungGrid.Tim(fmt.Sprintf(`button[data-team-id="%d"]`, idNhom))
				if nut.HopLe() {
					nut.GanSuKien("click", func(this js.Value, args []js.Value) any {
						xacNhan := js.Global().Call("confirm", fmt.Sprintf("Bạn có chắc chắn muốn xóa nhóm '%s'?\nCác dự án và thành viên thuộc nhóm sẽ được bỏ liên kết.", tenNhom))
						if !xacNhan.Bool() {
							return nil
						}

						nut.DatThuocTinh("disabled", "true")
						go func() {
							err := api.XoaNhom(idNhom)
							if err != nil {
								js.Global().Call("alert", "Lỗi: "+err.Error())
								nut.XoaThuocTinh("disabled")
								return
							}

							// Tải lại danh sách nhóm
							taiDanhSachNhom(khungNoiDung)
						}()
						return nil
					})
				}
			}
		}
	}()
}

/**
 * Hiển thị màn hình Chi tiết Nhóm làm việc (`#/teams/{id}`).
 * Bao gồm danh sách thành viên, form thêm thành viên vào nhóm (cho Admin & Team Lead),
 * và danh sách các dự án CI/CD trực thuộc nhóm.
 */
func HienThiTrangChiTietNhom(idChuoi string) {
	if !state.LayTrangThai().DaDangNhap() {
		routing.ChuyenHuong("#/login")
		return
	}

	id, err := strconv.ParseInt(idChuoi, 10, 64)
	if err != nil || id <= 0 {
		routing.ChuyenHuong("#/teams")
		return
	}

	khungChinh := dom.LayPhanTuTheoId("app")
	khungChinh.XoaHetCon()

	boKhung := dom.TaoPhanTu("div").ThemLopCss("dashboard-layout")
	thanhNav := TaoThanhDieuHuong("teams")
	boKhung.ThemCon(thanhNav)

	khungNoiDung := dom.TaoPhanTu("main").ThemLopCss("app-content")
	khungNoiDung.DatId("team-detail-content")

	khungNoiDung.DatHtml(`
		<div class="content-loading">
			<div class="spinner"></div>
			<h3>Đang tải thông tin chi tiết nhóm...</h3>
		</div>
	`)

	boKhung.ThemCon(khungNoiDung)
	khungChinh.ThemCon(boKhung)

	taiDuLieuChiTietNhom(khungNoiDung, id)
}

/**
 * Nạp chi tiết nhóm và thành viên từ API.
 */
func taiDuLieuChiTietNhom(khungNoiDung dom.PhanTu, nhomId int64) {
	go func() {
		nhom, err := api.LayChiTietNhom(nhomId)
		if err != nil {
			khungNoiDung.DatHtml(fmt.Sprintf(`
				<div class="card card-error">
					<h3>⚠️ Không thể tải thông tin nhóm</h3>
					<p>%s</p>
					<a href="#/teams" class="btn btn-outline" style="margin-top: 12px;">← Quay lại danh sách nhóm</a>
				</div>
			`, err.Error()))
			return
		}

		laQuanTri := state.LayTrangThai().LaTruongNhom()

		moTa := nhom.Description
		if moTa == "" {
			moTa = "Chưa có mô tả cho nhóm này."
		}

		nguoiTao := "Quản trị viên"
		if nhom.NguoiTao != nil && nhom.NguoiTao.Name != "" {
			nguoiTao = nhom.NguoiTao.Name
		}

		nutThemThanhVienHtml := ""
		if laQuanTri {
			nutThemThanhVienHtml = `
				<button id="btn-mo-form-them-tv" class="btn btn-primary btn-sm">
					<span>➕</span> Thêm Thành Viên
				</button>
			`
		}

		khungNoiDung.DatHtml(fmt.Sprintf(`
			<!-- Breadcrumb -->
			<div class="breadcrumb-nav">
				<a href="#/teams" class="back-link">← Quay lại danh sách nhóm</a>
				<span class="sep">/</span>
				<span class="curr-page">%s</span>
			</div>

			<!-- Header thông tin nhóm -->
			<div class="card team-info-banner">
				<div class="team-banner-left">
					<div class="team-banner-avatar">👥</div>
					<div class="team-banner-meta">
						<h1 class="team-banner-title">%s</h1>
						<p class="team-banner-desc">%s</p>
						<div class="team-banner-sub">
							<span>👤 Người tạo: <strong>%s</strong></span>
							<span class="dot-sep">•</span>
							<span>📅 Ngày tạo: <strong>%s</strong></span>
						</div>
					</div>
				</div>
			</div>

			<!-- Khung Form thêm thành viên (Ẩn mặc định) -->
			<div id="box-them-thanh-vien" class="card modal-inline-card" style="display: none; margin-bottom: 24px;">
				<div class="card-header-flex">
					<h3 style="font-size: 1.1rem; color: #f8fafc;">Thêm Thành Viên Vào Nhóm</h3>
					<button id="btn-dong-form-them-tv" class="btn btn-outline btn-xs">✖ Đóng</button>
				</div>
				<div style="margin-top: 14px;">
					<div class="form-row-2">
						<div class="form-group">
							<label class="form-label" for="sel-chon-user">Chọn người dùng <span class="required">*</span></label>
							<select id="sel-chon-user" class="form-control">
								<option value="">-- Đang tải danh sách người dùng... --</option>
							</select>
						</div>
						<div class="form-group">
							<label class="form-label" for="sel-chon-vai-tro">Vai trò trong nhóm <span class="required">*</span></label>
							<select id="sel-chon-vai-tro" class="form-control">
								<option value="viewer">Viewer (Chỉ xem và theo dõi pipeline)</option>
								<option value="team_lead">Team Lead (Trưởng nhóm, quản trị dự án & đóng cảnh báo)</option>
							</select>
						</div>
					</div>
					<div id="loi-them-thanh-vien" class="alert-error-box" style="display: none; margin-top: 10px;"></div>
					<div class="form-actions" style="margin-top: 14px; display: flex; gap: 10px;">
						<button id="btn-xac-nhan-them-tv" class="btn btn-primary btn-sm">
							<span>💾</span> Lưu Thành Viên
						</button>
						<button id="btn-huy-them-tv" class="btn btn-outline btn-sm">Hủy</button>
					</div>
				</div>
			</div>

			<!-- Bố cục 2 phần: Bảng Thành Viên & Dự án Trực Thuộc -->
			<div class="team-sections-grid">
				<!-- Cột 1: Danh sách Thành viên -->
				<div class="card team-members-section">
					<div class="section-title-bar">
						<div>
							<h3 class="section-title">Thành viên trong nhóm (%d)</h3>
							<p class="section-subtitle">Danh sách nhân sự và quyền hạn vận hành</p>
						</div>
						<div>
							%s
						</div>
					</div>

					<div class="table-responsive" style="margin-top: 16px;">
						<table class="data-table members-table">
							<thead>
								<tr>
									<th>Thành viên</th>
									<th>Email</th>
									<th>Vai trò nhóm</th>
									<th style="text-align: right;">Thao tác</th>
								</tr>
							</thead>
							<tbody id="tbody-danh-sach-thanh-vien">
								<!-- Nạp hàng thành viên -->
							</tbody>
						</table>
					</div>
				</div>

				<!-- Cột 2: Dự án Trực thuộc nhóm -->
				<div class="card team-projects-section">
					<div class="section-title-bar">
						<div>
							<h3 class="section-title">Dự án CI/CD Trực thuộc (%d)</h3>
							<p class="section-subtitle">Các dự án do nhóm trực tiếp quản lý và giám sát</p>
						</div>
						<a href="#/projects" class="btn btn-outline btn-xs">Tất cả dự án →</a>
					</div>

					<div id="team-projects-list" style="margin-top: 16px;">
						<!-- Nạp danh sách dự án -->
					</div>
				</div>
			</div>
		`, nhom.Name, nhom.Name, moTa, nguoiTao, nhom.CreatedAt,
			len(nhom.DanhSachThanhVien), nutThemThanhVienHtml, len(nhom.DanhSachDuAn)))

		// Render bảng thành viên
		tbodyThanhVien := khungNoiDung.Tim("#tbody-danh-sach-thanh-vien")
		if tbodyThanhVien.HopLe() {
			if len(nhom.DanhSachThanhVien) == 0 {
				tbodyThanhVien.DatHtml(`
					<tr>
						<td colspan="4" class="text-center text-muted" style="padding: 24px;">
							Chưa có thành viên nào trong nhóm này.
						</td>
					</tr>
				`)
			} else {
				var sbTv strings.Builder
				for _, tv := range nhom.DanhSachThanhVien {
					tenVaiTro := "Viewer"
					lopVaiTro := "badge-role-viewer"
					if tv.Role == "team_lead" {
						tenVaiTro = "Team Lead"
						lopVaiTro = "badge-role-lead"
					}

					nutXoaTv := ""
					if laQuanTri {
						nutXoaTv = fmt.Sprintf(`
							<button class="btn btn-outline btn-xs btn-danger-hover btn-xoa-thanh-vien" data-user-id="%d" data-user-name="%s" title="Xóa khỏi nhóm">
								🗑️ Xóa
							</button>
						`, tv.ID, tv.Name)
					}

					sbTv.WriteString(fmt.Sprintf(`
						<tr>
							<td class="user-cell">
								<div class="table-user-chip">
									<div class="table-user-avatar">%s</div>
									<span class="table-user-name">%s</span>
								</div>
							</td>
							<td class="text-muted">%s</td>
							<td>
								<span class="badge %s">%s</span>
							</td>
							<td style="text-align: right;">
								%s
							</td>
						</tr>
					`, strings.ToUpper(string([]rune(tv.Name)[0])), tv.Name, tv.Email, lopVaiTro, tenVaiTro, nutXoaTv))
				}
				tbodyThanhVien.DatHtml(sbTv.String())
			}
		}

		// Render danh sách dự án trực thuộc
		khungDuAn := khungNoiDung.Tim("#team-projects-list")
		if khungDuAn.HopLe() {
			if len(nhom.DanhSachDuAn) == 0 {
				khungDuAn.DatHtml(`
					<div class="empty-state-mini">
						<p class="text-muted">Nhóm chưa có dự án CI/CD nào được liên kết.</p>
					</div>
				`)
			} else {
				var sbDa strings.Builder
				sbDa.WriteString(`<div class="team-projects-grid">`)
				for _, da := range nhom.DanhSachDuAn {
					sbDa.WriteString(fmt.Sprintf(`
						<div class="team-project-item">
							<div class="tp-header">
								<h4 class="tp-title"><a href="#/projects/%d">📁 %s</a></h4>
								<span class="badge badge-provider">%s</span>
							</div>
							<p class="tp-repo"><code>%s</code></p>
							<div class="tp-footer">
								<span class="tp-threshold">Ngưỡng lỗi: %d lần</span>
								<a href="#/projects/%d" class="btn btn-outline btn-xs">Xem chi tiết →</a>
							</div>
						</div>
					`, da.ID, da.Name, strings.ToUpper(da.CIProvider), da.RepositoryURL, da.FailureThreshold, da.ID))
				}
				sbDa.WriteString(`</div>`)
				khungDuAn.DatHtml(sbDa.String())
			}
		}

		// Xử lý Form thêm thành viên (nếu có quyền)
		if laQuanTri {
			boxThemTv := khungNoiDung.Tim("#box-them-thanh-vien")
			btnMoThemTv := khungNoiDung.Tim("#btn-mo-form-them-tv")
			btnDongThemTv := khungNoiDung.Tim("#btn-dong-form-them-tv")
			btnHuyThemTv := khungNoiDung.Tim("#btn-huy-them-tv")
			selChonUser := khungNoiDung.Tim("#sel-chon-user")

			// Khi mở form, gọi API lấy danh sách người dùng để đưa vào select
			if btnMoThemTv.HopLe() && boxThemTv.HopLe() {
				btnMoThemTv.GanSuKien("click", func(this js.Value, args []js.Value) any {
					boxThemTv.DatThuocTinh("style", "display: block; margin-bottom: 24px;")

					// Nạp danh sách người dùng cho dropdown
					go func() {
						users, errUsers := api.LayDanhSachNguoiDung()
						if errUsers != nil || len(users) == 0 {
							if selChonUser.HopLe() {
								selChonUser.DatHtml(`<option value="">(Không thể tải danh sách người dùng)</option>`)
							}
							return
						}

						var optSb strings.Builder
						optSb.WriteString(`<option value="">-- Chọn tài khoản người dùng --</option>`)
						for _, u := range users {
							optSb.WriteString(fmt.Sprintf(`<option value="%d">%s (%s - %s)</option>`, u.ID, u.Name, u.Email, u.Role))
						}
						if selChonUser.HopLe() {
							selChonUser.DatHtml(optSb.String())
						}
					}()
					return nil
				})
			}

			dongBoxThemTv := func() {
				if boxThemTv.HopLe() {
					boxThemTv.DatThuocTinh("style", "display: none; margin-bottom: 24px;")
				}
			}

			if btnDongThemTv.HopLe() {
				btnDongThemTv.GanSuKien("click", func(this js.Value, args []js.Value) any {
					dongBoxThemTv()
					return nil
				})
			}
			if btnHuyThemTv.HopLe() {
				btnHuyThemTv.GanSuKien("click", func(this js.Value, args []js.Value) any {
					dongBoxThemTv()
					return nil
				})
			}

			// Nút Xác nhận Lưu Thành Viên
			btnLuuTv := khungNoiDung.Tim("#btn-xac-nhan-them-tv")
			if btnLuuTv.HopLe() {
				btnLuuTv.GanSuKien("click", func(this js.Value, args []js.Value) any {
					selVaiTro := khungNoiDung.Tim("#sel-chon-vai-tro")
					hopLoiTv := khungNoiDung.Tim("#loi-them-thanh-vien")

					valUserStr := selChonUser.LayGiaTri()
					valVaiTro := selVaiTro.LayGiaTri()

					userId, errP := strconv.ParseInt(valUserStr, 10, 64)
					if errP != nil || userId <= 0 {
						if hopLoiTv.HopLe() {
							hopLoiTv.DatThuocTinh("style", "display: block; margin-top: 10px;")
							hopLoiTv.DatHtml("Vui lòng chọn một người dùng hợp lệ.")
						}
						return nil
					}

					if hopLoiTv.HopLe() {
						hopLoiTv.DatThuocTinh("style", "display: none;")
					}

					btnLuuTv.DatThuocTinh("disabled", "true")
					btnLuuTv.DatHtml("Đang lưu...")

					go func() {
						errSave := api.ThemThanhVienVaoNhom(nhomId, userId, valVaiTro)
						btnLuuTv.XoaThuocTinh("disabled")
						btnLuuTv.DatHtml("<span>💾</span> Lưu Thành Viên")

						if errSave != nil {
							if hopLoiTv.HopLe() {
								hopLoiTv.DatThuocTinh("style", "display: block; margin-top: 10px;")
								hopLoiTv.DatHtml("Lỗi: " + errSave.Error())
							}
							return
						}

						// Thành công -> đóng form và nạp lại dữ liệu nhóm
						dongBoxThemTv()
						taiDuLieuChiTietNhom(khungNoiDung, nhomId)
					}()
					return nil
				})
			}

			// Gán sự kiện Xóa thành viên
			for _, tv := range nhom.DanhSachThanhVien {
				idTv := tv.ID
				tenTv := tv.Name
				nutXoa := tbodyThanhVien.Tim(fmt.Sprintf(`button[data-user-id="%d"]`, idTv))
				if nutXoa.HopLe() {
					nutXoa.GanSuKien("click", func(this js.Value, args []js.Value) any {
						xacNhan := js.Global().Call("confirm", fmt.Sprintf("Bạn có chắc muốn xóa thành viên '%s' khỏi nhóm này?", tenTv))
						if !xacNhan.Bool() {
							return nil
						}

						nutXoa.DatThuocTinh("disabled", "true")
						go func() {
							errDel := api.XoaThanhVienKhoiNhom(nhomId, idTv)
							if errDel != nil {
								js.Global().Call("alert", "Lỗi: "+errDel.Error())
								nutXoa.XoaThuocTinh("disabled")
								return
							}

							// Nạp lại chi tiết nhóm
							taiDuLieuChiTietNhom(khungNoiDung, nhomId)
						}()
						return nil
					})
				}
			}
		}
	}()
}
