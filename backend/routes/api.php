<?php

use App\Http\Controllers\XacThucController;
use Illuminate\Support\Facades\Route;

/**
 * Định nghĩa các tuyến đường API cho hệ thống DevOps Sentinel (Phiên bản v1).
 */
Route::prefix('v1')->group(function () {

    // Tuyến đường kiểm tra trạng thái hoạt động của hệ thống
    Route::get('/health', function () {
        return response()->json([
            'success' => true,
            'message' => 'DevOps Sentinel API đang hoạt động bình thường.',
            'version' => '1.0.0-phase4',
            'timestamp' => now()->toIso8601String(),
        ]);
    });

    // Nhóm API Xác thực người dùng (Authentication)
    Route::prefix('auth')->group(function () {
        // Tuyến công khai (Public)
        Route::post('/register', [XacThucController::class, 'dangKy']);
        Route::post('/login', [XacThucController::class, 'dangNhap']);
        Route::post('/refresh', [XacThucController::class, 'lamMoiToken']);

        // Tuyến yêu cầu đăng nhập và có Access Token hợp lệ
        Route::middleware('jwt')->group(function () {
            Route::get('/me', [XacThucController::class, 'thongTinCaNhan']);
            Route::post('/logout', [XacThucController::class, 'dangXuat']);
        });
    });

    // Nhóm API kiểm tra phân quyền vai trò (RBAC Testing)
    Route::middleware('jwt')->group(function () {
        // Chỉ dành cho Quản trị viên tối cao (Admin)
        Route::get('/admin/kiem-tra-quyen', function () {
            return response()->json([
                'success' => true,
                'message' => 'Chào mừng Quản trị viên! Bạn có toàn quyền truy cập khu vực này.',
            ]);
        })->middleware('role:admin');

        // Dành cho Trưởng nhóm trở lên (Admin và Team Lead)
        Route::get('/lead/kiem-tra-quyen', function () {
            return response()->json([
                'success' => true,
                'message' => 'Chào mừng Trưởng nhóm! Bạn có quyền quản lý dự án và thành viên nhóm.',
            ]);
        })->middleware('role:team_lead');
    });
});
