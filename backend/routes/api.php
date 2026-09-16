<?php

use Illuminate\Support\Facades\Route;

/**
 * Tuyến đường kiểm tra sức khỏe của API Backend.
 * Mục đích: Đảm bảo nền tảng Backend hoạt động và kết nối tốt với cơ sở dữ liệu.
 */
Route::prefix('v1')->group(function () {
    Route::get('/health', function () {
        return response()->json([
            'success' => true,
            'message' => 'DevOps Sentinel API đang hoạt động bình thường.',
            'version' => '1.0.0-phase3',
            'timestamp' => now()->toIso8601String(),
        ]);
    });
});
