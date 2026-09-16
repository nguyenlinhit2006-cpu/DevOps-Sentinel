<?php

namespace App\Http\Controllers;

use App\Models\DuAn;
use App\Models\LuotChayPipeline;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

/**
 * Controller quản lý và truy xuất lịch sử các lượt chạy pipeline CI/CD (Pipeline Runs).
 */
class LuotChayController extends Controller
{
    /**
     * Lấy danh sách các lượt chạy của một dự án kèm bộ lọc (nhánh, trạng thái) và phân trang.
     */
    public function danhSachTheoDuAn(Request $yeuCau, int $duAnId): JsonResponse
    {
        $duAn = DuAn::find($duAnId);
        if (!$duAn) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy dự án.'],
            ], 404);
        }

        $tenNhanh = $yeuCau->query('branch');
        $trangThai = $yeuCau->query('status');
        $soLuongMoiTrang = (int) $yeuCau->query('per_page', 20);

        $truyVan = LuotChayPipeline::where('project_id', $duAnId);

        if ($tenNhanh) {
            $truyVan->where('branch', $tenNhanh);
        }

        if ($trangThai) {
            $truyVan->where('status', $trangThai);
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

    /**
     * Lấy thông tin chi tiết của một lượt chạy cụ thể (Bao gồm log tóm tắt).
     */
    public function chiTiet(int $id): JsonResponse
    {
        $luotChay = LuotChayPipeline::with('duAn:id,name,slug,ci_provider')->find($id);

        if (!$luotChay) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy lượt chạy pipeline yêu cầu.'],
            ], 404);
        }

        return response()->json([
            'success' => true,
            'data' => $luotChay,
        ]);
    }
}
