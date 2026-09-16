<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Khởi tạo bảng lịch sử thực thi pipeline (pipeline_runs).
     * Mục đích: Lưu trữ toàn bộ kết quả các lần build/test từ webhook do CI/CD provider đẩy về.
     */
    public function up(): void
    {
        Schema::create('pipeline_runs', function (Blueprint $table) {
            $table->id();
            $table->foreignId('project_id')->constrained('projects')->cascadeOnDelete();
            $table->string('pipeline_name', 100)->default('default'); // Tên workflow hoặc job
            $table->string('commit_hash', 64);
            $table->string('commit_message', 255)->nullable();
            $table->string('branch', 100);
            $table->string('author', 100)->nullable();
            $table->string('status', 20); // Trạng thái: queued, running, success, failed, cancelled
            $table->string('trigger_event', 50)->nullable(); // Sự kiện kích hoạt: push, pull_request, manual, schedule
            $table->integer('duration_seconds')->default(0); // Thời gian thực thi tính bằng giây
            $table->text('short_log')->nullable(); // Trích đoạn log tóm tắt hoặc thông báo lỗi (giới hạn 64KB)
            $table->string('external_url', 500)->nullable(); // Đường dẫn trực tiếp đến trang build của CI Provider
            $table->timestamp('started_at')->nullable();
            $table->timestamp('finished_at')->nullable();
            $table->timestamp('created_at')->useCurrent();

            // Các chỉ mục tối ưu hóa tốc độ truy vấn dashboard và kiểm tra lỗi liên tiếp
            $table->index(['project_id', 'created_at']);
            $table->index(['project_id', 'status']);
            $table->index('commit_hash');
        });
    }

    /**
     * Thu hồi bảng pipeline_runs khi rollback.
     */
    public function down(): void
    {
        Schema::dropIfExists('pipeline_runs');
    }
};
