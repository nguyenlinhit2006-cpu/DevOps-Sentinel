<?php

namespace App\Services;

use App\Models\NguoiDung;
use App\Models\NhatKyHoatDong;
use Illuminate\Http\Request;

/**
 * Lớp dịch vụ ghi nhật ký kiểm toán hệ thống (Audit Log).
 * Phục vụ giám sát an toàn thông tin và truy vết các thao tác quan trọng.
 */
class DichVuAuditLog
{
    /**
     * Ghi một bản ghi nhật ký kiểm toán vào cơ sở dữ liệu.
     *
     * @param string $hanhDong Tên hành động (VD: tao_du_an, cap_nhat_du_an, xoa_du_an, xu_ly_webhook...)
     * @param string $loaiThucThe Loại đối tượng thao tác (project, team, alert, user, system)
     * @param int|null $idThucThe Khóa chính của đối tượng
     * @param array|null $chiTiet Dữ liệu chi tiết bổ trợ dạng mảng (sẽ lưu dưới dạng JSONB)
     * @param NguoiDung|null $nguoiThaoTac Người thực hiện (nếu null sẽ lấy từ request hiện tại)
     * @param Request|null $yeuCau Request HTTP đi kèm để lấy IP và User-Agent
     */
    public function ghiNhatKy(
        string $hanhDong,
        string $loaiThucThe,
        ?int $idThucThe = null,
        ?array $chiTiet = null,
        ?NguoiDung $nguoiThaoTac = null,
        ?Request $yeuCau = null
    ): NhatKyHoatDong {
        $yeuCauHienTai = $yeuCau ?? request();

        $nguoiDungId = $nguoiThaoTac?->id
            ?? $yeuCauHienTai?->attributes->get('nguoi_dung_hien_tai')?->id
            ?? null;

        $diaChiIp = $yeuCauHienTai?->ip();
        $thongTinTrinhDuyet = substr($yeuCauHienTai?->userAgent() ?? 'Unknown', 0, 255);

        return NhatKyHoatDong::create([
            'user_id' => $nguoiDungId,
            'action' => $hanhDong,
            'entity_type' => $loaiThucThe,
            'entity_id' => $idThucThe,
            'ip_address' => $diaChiIp,
            'user_agent' => $thongTinTrinhDuyet,
            'details' => $chiTiet,
            'created_at' => now(),
        ]);
    }
}
