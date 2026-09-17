//go:build js && wasm

package state

import (
	"time"
)

/**
 * Khởi chạy bộ polling ngầm thời gian thực (định kỳ 4 giây theo thỏa thuận Phase 0).
 * Nhận vào hàm callback xử lý cập nhật tương ứng với tuyến đường hiện tại của người dùng.
 */
func KhoiDongPollingRealtime(hamCapNhat func(duongDanHienTai string)) {
	go func() {
		conDem := time.NewTicker(4 * time.Second)
		defer conDem.Stop()

		for range conDem.C {
			tt := LayTrangThai()
			if !tt.DaDangNhap() || !tt.PollingKichHoat {
				continue
			}

			if hamCapNhat != nil {
				hamCapNhat("")
			}
		}
	}()
}
