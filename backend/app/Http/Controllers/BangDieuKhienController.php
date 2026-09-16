<?php

namespace App\Http\Controllers;

use App\Models\CanhBao;
use App\Models\DuAn;
use App\Models\LuotChayPipeline;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\DB;

/**
 * Controller cung cấp các API tổng quan và số liệu thống kê cho Dashboard thời gian thực.
 */
class BangDieuKhienController extends Controller
{
    /**
     * Lấy các chỉ số tóm tắt toàn hệ thống (Tổng dự án, tỷ lệ thành công, cảnh báo active, lượt chạy gần nhất).
     */
    public function tongQuan(): JsonResponse
    {
        $tongSoDuAn = DuAn::count();
        $tongSoLuotChay = LuotChayPipeline::count();
        $soLuotThanhCong = LuotChayPipeline::where('status', 'success')->count();
        $soLuotThatBai = LuotChayPipeline::where('status', 'failed')->count();

        $homNay = Carbon::today();
        $soLuotChayHomNay = LuotChayPipeline::where('created_at', '>=', $homNay)->count();

        $tyLeThanhCong = $tongSoLuotChay > 0
            ? round(($soLuotThanhCong / $tongSoLuotChay) * 100, 2)
            : 100.0;

        $soCanhBaoDangMo = CanhBao::where('status', 'active')->count();

        // 8 lượt chạy gần đây nhất
        $danhSachLuotChayMoiNhat = LuotChayPipeline::with('duAn:id,name,slug')
            ->orderBy('id', 'desc')
            ->limit(8)
            ->get();

        return response()->json([
            'success' => true,
            'data' => [
                'total_projects' => $tongSoDuAn,
                'total_runs' => $tongSoLuotChay,
                'total_runs_today' => $soLuotChayHomNay,
                'success_runs_count' => $soLuotThanhCong,
                'failed_runs_count' => $soLuotThatBai,
                'success_rate_percent' => $tyLeThanhCong,
                'active_alerts_count' => $soCanhBaoDangMo,
                'recent_runs' => $danhSachLuotChayMoiNhat,
            ],
        ]);
    }

    /**
     * Lấy chuỗi số liệu thống kê theo ngày (Aggregate Query) phục vụ vẽ biểu đồ tỷ lệ thành công/thất bại.
     */
    public function soLieuThongKe(Request $yeuCau): JsonResponse
    {
        $soNgay = min((int) $yeuCau->query('days', 7), 30);
        $ngayBatDau = Carbon::now()->subDays($soNgay - 1)->startOfDay();

        // Truy vấn tổng hợp gom nhóm theo ngày
        $ketQuaGomNhom = LuotChayPipeline::select([
            DB::raw("DATE(created_at) as date_string"),
            DB::raw("COUNT(CASE WHEN status = 'success' THEN 1 END) as success_count"),
            DB::raw("COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count"),
            DB::raw("COUNT(*) as total_count"),
        ])
        ->where('created_at', '>=', $ngayBatDau)
        ->groupBy(DB::raw("DATE(created_at)"))
        ->orderBy('date_string', 'asc')
        ->get();

        // Tạo mảng dữ liệu liên tục cho từng ngày (ngay cả những ngày không có build)
        $chuoiThoiGian = [];
        $ngayLap = $ngayBatDau->copy();
        $ngayKetThuc = Carbon::now()->startOfDay();

        $bangTraCuu = $ketQuaGomNhom->keyBy('date_string');

        while ($ngayLap->lte($ngayKetThuc)) {
            $dinhDangNgay = $ngayLap->format('Y-m-d');
            $duLieuNgay = $bangTraCuu->get($dinhDangNgay);

            $chuoiThoiGian[] = [
                'date' => $dinhDangNgay,
                'success_count' => (int) ($duLieuNgay->success_count ?? 0),
                'failed_count' => (int) ($duLieuNgay->failed_count ?? 0),
                'total_count' => (int) ($duLieuNgay->total_count ?? 0),
            ];

            $ngayLap->addDay();
        }

        return response()->json([
            'success' => true,
            'data' => [
                'days' => $soNgay,
                'daily_metrics' => $chuoiThoiGian,
            ],
        ]);
    }
}
