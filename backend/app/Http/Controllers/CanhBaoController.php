<?php

namespace App\Http\Controllers;

use App\Models\CanhBao;
use App\Models\NguoiDung;
use App\Services\DichVuAuditLog;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

/**
 * Controller quản lý danh sách và xử lý đóng Cảnh báo (Alerts).
 */
class CanhBaoController extends Controller
{
    protected DichVuAuditLog $dichVuAudit;

    public function __construct(DichVuAuditLog $dichVuAudit)
    {
        $this->dichVuAudit = $dichVuAudit;
    }

    /**
     * Lấy danh sách các cảnh báo theo bộ lọc (active/resolved) và dự án.
     */
    public function danhSach(Request $yeuCau): JsonResponse
    {
        $trangThai = $yeuCau->query('status');
        $idDuAn = $yeuCau->query('project_id');

        $truyVan = CanhBao::with([
            'duAn:id,name,slug',
            'luotChayLoi:id,branch,commit_hash,status,created_at',
            'nguoiXuLy:id,name,email',
        ]);

        if ($trangThai) {
            $truyVan->where('status', $trangThai);
        }

        if ($idDuAn) {
            $truyVan->where('project_id', $idDuAn);
        }

        $danhSachCanhBao = $truyVan->orderBy('created_at', 'desc')->get();

        return response()->json([
            'success' => true,
            'data' => $danhSachCanhBao,
        ]);
    }

    /**
     * Xác nhận đóng/xử lý cảnh báo (Yêu cầu vai trò Team Lead trở lên).
     */
    public function dongCanhBao(Request $yeuCau, int $id): JsonResponse
    {
        $canhBao = CanhBao::find($id);

        if (!$canhBao) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy cảnh báo yêu cầu.'],
            ], 404);
        }

        /** @var NguoiDung $nguoiDung */
        $nguoiDung = $yeuCau->attributes->get('nguoi_dung_hien_tai');

        $canhBao->update([
            'status' => 'resolved',
            'resolved_by' => $nguoiDung->id,
            'resolved_at' => now(),
        ]);

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'dong_canh_bao',
            loaiThucThe: 'alert',
            idThucThe: $canhBao->id,
            chiTiet: ['project_id' => $canhBao->project_id, 'title' => $canhBao->title]
        );

        return response()->json([
            'success' => true,
            'data' => $canhBao,
            'message' => 'Đã đóng và xử lý cảnh báo thành công.',
        ]);
    }
}
