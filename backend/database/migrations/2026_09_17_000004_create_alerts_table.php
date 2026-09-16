<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Khởi tạo bảng cảnh báo (alerts).
     * Mục đích: Quản lý các cảnh báo được kích hoạt tự động khi một pipeline thất bại liên tiếp N lần.
     */
    public function up(): void
    {
        Schema::create('alerts', function (Blueprint $table) {
            $table->id();
            $table->foreignId('project_id')->constrained('projects')->cascadeOnDelete();
            $table->foreignId('last_pipeline_run_id')->nullable()->constrained('pipeline_runs')->nullOnDelete();
            $table->string('title', 255);
            $table->text('message');
            $table->integer('consecutive_failures')->default(1); // Ghi nhận số lần lỗi liên tiếp thực tế
            $table->string('status', 20)->default('active'); // Trạng thái: active, acknowledged, resolved
            $table->foreignId('resolved_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamp('resolved_at')->nullable();
            $table->timestamps();

            // Chỉ mục lọc cảnh báo đang kích hoạt và theo thời gian
            $table->index(['project_id', 'status']);
            $table->index(['status', 'created_at']);
        });
    }

    /**
     * Thu hồi bảng alerts khi rollback.
     */
    public function down(): void
    {
        Schema::dropIfExists('alerts');
    }
};
