<?php

namespace App\Http\Controllers;

use App\Models\NhatKyHoatDong;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

/**
 * Controller truy vấn nhật ký kiểm toán hệ thống (Audit Logs) dành cho Quản trị viên.
 */
class NhatKyController extends Controller
{
    /**
     * Lấy danh sách nhật ký thao tác hệ thống kèm bộ lọc và phân trang.
     */
    public function danhSach(Request $yeuCau): JsonResponse
    {
        $idNguoiDung = $yeuCau->query('user_id');
        $tenHanhDong = $yeuCau->query('action');
        $loaiThucThe = $yeuCau->query('entity_type');
        $soLuongMoiTrang = (int) $yeuCau->query('per_page', 25);

        $truyVan = NhatKyHoatDong::with('nguoiThaoTac:id,name,email,role');

        if ($idNguoiDung) {
            $truyVan->where('user_id', $idNguoiDung);
        }

        if ($tenHanhDong) {
            $truyVan->where('action', 'like', "%{$tenHanhDong}%");
        }

        if ($loaiThucThe) {
            $truyVan->where('entity_type', $loaiThucThe);
        }

        $ketQuaPhanTrang = $truyVan->orderBy('id', 'desc')->paginate($soLuongMoiTrang);

        return response()->json([
            'success' => true,
            'data' => $ketQuaPhanTrang->items(),
            'meta' => [
                'page' => $ketQuaPhanTrang->currentPage(),
                'per_page' => $ketQuaPhanTrang->perPage(),
                'total' => $ketQuaPhanTrang->total(),
                'total_pages' => $ketQuaPhanTrang->lastPage(),
            ],
        ]);
    }
}
