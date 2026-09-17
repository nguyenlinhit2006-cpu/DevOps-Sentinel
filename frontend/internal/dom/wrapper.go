//go:build js && wasm

package dom

import (
	"syscall/js"
)

/**
 * Kiểu dữ liệu đại diện cho một phần tử DOM trong trình duyệt.
 * Bao đóng đối tượng js.Value để thao tác DOM trực tiếp, an toàn và dễ tái sử dụng.
 */
type PhanTu struct {
	GiaTri js.Value
}

/**
 * Lấy đối tượng tài liệu toàn cục (document) của trình duyệt.
 */
func LayTaiLieu() PhanTu {
	return PhanTu{GiaTri: js.Global().Get("document")}
}

/**
 * Lấy đối tượng cửa sổ toàn cục (window) của trình duyệt.
 */
func LayCuaSo() PhanTu {
	return PhanTu{GiaTri: js.Global()}
}

/**
 * Tìm kiếm phần tử DOM theo định danh ID.
 */
func LayPhanTuTheoId(id string) PhanTu {
	giaTri := LayTaiLieu().GiaTri.Call("getElementById", id)
	return PhanTu{GiaTri: giaTri}
}

/**
 * Tạo một phần tử HTML mới theo tên thẻ (vd: div, button, input...).
 */
func TaoPhanTu(tenThe string) PhanTu {
	giaTri := LayTaiLieu().GiaTri.Call("createElement", tenThe)
	return PhanTu{GiaTri: giaTri}
}

/**
 * Kiểm tra xem phần tử DOM có tồn tại và hợp lệ hay không.
 */
func (pt PhanTu) HopLe() bool {
	return !pt.GiaTri.IsNull() && !pt.GiaTri.IsUndefined()
}

/**
 * Gán thuộc tính (attribute) cho phần tử HTML.
 */
func (pt PhanTu) DatThuocTinh(tenThuocTinh, giaTri string) PhanTu {
	if pt.HopLe() {
		pt.GiaTri.Call("setAttribute", tenThuocTinh, giaTri)
	}
	return pt
}

/**
 * Đặt nội dung văn bản thuần (innerText) cho phần tử.
 */
func (pt PhanTu) DatNoiDung(noiDung string) PhanTu {
	if pt.HopLe() {
		pt.GiaTri.Set("innerText", noiDung)
	}
	return pt
}

/**
 * Đặt nội dung HTML (innerHTML) cho phần tử.
 */
func (pt PhanTu) DatHtml(noiDungHtml string) PhanTu {
	if pt.HopLe() {
		pt.GiaTri.Set("innerHTML", noiDungHtml)
	}
	return pt
}

/**
 * Thêm một hoặc nhiều phần tử con vào trong phần tử hiện tại.
 */
func (pt PhanTu) ThemCon(danhSachCon ...PhanTu) PhanTu {
	if pt.HopLe() {
		for _, con := range danhSachCon {
			if con.HopLe() {
				pt.GiaTri.Call("appendChild", con.GiaTri)
			}
		}
	}
	return pt
}

/**
 * Xóa sạch toàn bộ phần tử con bên trong.
 */
func (pt PhanTu) XoaHetCon() PhanTu {
	if pt.HopLe() {
		pt.GiaTri.Set("innerHTML", "")
	}
	return pt
}

/**
 * Lấy giá trị chuỗi (value) của phần tử form input.
 */
func (pt PhanTu) LayGiaTri() string {
	if pt.HopLe() {
		return pt.GiaTri.Get("value").String()
	}
	return ""
}

/**
 * Đặt giá trị chuỗi (value) cho phần tử form input.
 */
func (pt PhanTu) DatGiaTri(giaTri string) PhanTu {
	if pt.HopLe() {
		pt.GiaTri.Set("value", giaTri)
	}
	return pt
}

/**
 * Thêm một hoặc nhiều lớp định kiểu CSS (CSS Classes) vào phần tử.
 */
func (pt PhanTu) ThemLopCss(danhSachLop ...string) PhanTu {
	if pt.HopLe() {
		classList := pt.GiaTri.Get("classList")
		for _, lop := range danhSachLop {
			if lop != "" {
				classList.Call("add", lop)
			}
		}
	}
	return pt
}

/**
 * Xóa lớp định kiểu CSS khỏi phần tử.
 */
func (pt PhanTu) XoaLopCss(danhSachLop ...string) PhanTu {
	if pt.HopLe() {
		classList := pt.GiaTri.Get("classList")
		for _, lop := range danhSachLop {
			if lop != "" {
				classList.Call("remove", lop)
			}
		}
	}
	return pt
}

/**
 * Lắng nghe và gắn sự kiện tương tác người dùng (click, submit, change...).
 */
func (pt PhanTu) GanSuKien(tenSuKien string, hamXuLy func(this js.Value, args []js.Value) any) js.Func {
	hamJs := js.FuncOf(hamXuLy)
	if pt.HopLe() {
		pt.GiaTri.Call("addEventListener", tenSuKien, hamJs)
	}
	return hamJs
}
