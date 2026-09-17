<?php

use App\Http\Controllers\BangDieuKhienController;
use App\Http\Controllers\CanhBaoController;
use App\Http\Controllers\DuAnController;
use App\Http\Controllers\LuotChayController;
use App\Http\Controllers\NhatKyController;
use App\Http\Controllers\NhomController;
use App\Http\Controllers\WebhookController;
use App\Http\Controllers\XacThucController;
use Illuminate\Support\Facades\Route;

/**
 * Định nghĩa toàn bộ hệ thống Tuyến đường RESTful API v1 cho DevOps Sentinel.
 */
Route::prefix('v1')->group(function () {

    // 1. Tuyến kiểm tra sức khỏe hệ thống
    Route::get('/health', function () {
        return response()->json([
            'success' => true,
            'message' => 'DevOps Sentinel API đang hoạt động bình thường.',
            'version' => '1.0.0-phase5',
            'timestamp' => now()->toIso8601String(),
        ]);
    });

    // 2. Nhóm xác thực người dùng (Public)
    Route::prefix('auth')->group(function () {
        Route::post('/register', [XacThucController::class, 'dangKy']);
        Route::post('/login', [XacThucController::class, 'dangNhap']);
        Route::post('/refresh', [XacThucController::class, 'lamMoiToken']);
    });

    // 3. Nhóm Webhook Ingestion (Bảo vệ bằng chữ ký HMAC SHA-256)
    Route::post('/webhooks/projects/{id}', [WebhookController::class, 'tiepNhan'])
        ->middleware('webhook.hmac');

    // 4. Toàn bộ các API nghiệp vụ cốt lõi (Bảo vệ bằng JWT và Audit Log Middleware)
    Route::middleware(['jwt', 'audit'])->group(function () {

        // Quản lý thông tin tài khoản & danh sách người dùng
        Route::get('/auth/me', [XacThucController::class, 'thongTinCaNhan']);
        Route::post('/auth/logout', [XacThucController::class, 'dangXuat']);
        Route::get('/users', function () {
            $danhSach = \App\Models\NguoiDung::select('id', 'name', 'email', 'role')->orderBy('id')->get();
            return response()->json([
                'success' => true,
                'data' => $danhSach,
            ]);
        });

        // Nhóm API Dự án (Projects)
        Route::prefix('projects')->group(function () {
            Route::get('/', [DuAnController::class, 'danhSach']);
            Route::post('/', [DuAnController::class, 'taoMoi'])->middleware('role:team_lead');
            Route::get('/{id}', [DuAnController::class, 'chiTiet']);
            Route::put('/{id}', [DuAnController::class, 'capNhat'])->middleware('role:team_lead');
            Route::delete('/{id}', [DuAnController::class, 'xoa'])->middleware('role:admin');
            Route::post('/{id}/regenerate-secret', [DuAnController::class, 'sinhLaiSecret'])->middleware('role:team_lead');

            // Lượt chạy pipeline theo dự án
            Route::get('/{projectId}/runs', [LuotChayController::class, 'danhSachTheoDuAn']);
        });

        // Chi tiết 1 lượt chạy pipeline
        Route::get('/runs/{id}', [LuotChayController::class, 'chiTiet']);

        // Nhóm API Nhóm làm việc (Teams)
        Route::prefix('teams')->group(function () {
            Route::get('/', [NhomController::class, 'danhSach']);
            Route::post('/', [NhomController::class, 'taoMoi'])->middleware('role:admin');
            Route::get('/{id}', [NhomController::class, 'chiTiet']);
            Route::put('/{id}', [NhomController::class, 'capNhat'])->middleware('role:team_lead');
            Route::delete('/{id}', [NhomController::class, 'xoa'])->middleware('role:admin');
            Route::post('/{id}/members', [NhomController::class, 'themThanhVien'])->middleware('role:team_lead');
            Route::delete('/{id}/members/{userId}', [NhomController::class, 'xoaThanhVien'])->middleware('role:team_lead');
        });

        // Nhóm API Dashboard & Metrics
        Route::prefix('dashboard')->group(function () {
            Route::get('/summary', [BangDieuKhienController::class, 'tongQuan']);
            Route::get('/metrics', [BangDieuKhienController::class, 'soLieuThongKe']);
        });

        // Nhóm API Cảnh báo Sentinel (Alerts)
        Route::prefix('alerts')->group(function () {
            Route::get('/', [CanhBaoController::class, 'danhSach']);
            Route::post('/{id}/resolve', [CanhBaoController::class, 'dongCanhBao'])->middleware('role:team_lead');
        });

        // Nhóm API Nhật ký Kiểm toán (Audit Logs - Chỉ Admin)
        Route::get('/audit-logs', [NhatKyController::class, 'danhSach'])->middleware('role:admin');

        // Tuyến kiểm tra nhanh quyền vai trò Admin
        Route::get('/admin/kiem-tra-quyen', function () {
            return response()->json([
                'success' => true,
                'message' => 'Chào mừng Quản trị viên! Bạn có toàn quyền truy cập.',
            ]);
        })->middleware('role:admin');
    });
});
