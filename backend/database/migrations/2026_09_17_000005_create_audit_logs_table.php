<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Khởi tạo bảng nhật ký kiểm toán (audit_logs).
     * Mục đích: Lưu vết toàn bộ các thao tác nhạy cảm (tạo dự án, đóng cảnh báo, phân quyền) phục vụ an toàn hệ thống.
     */
    public function up(): void
    {
        Schema::create('audit_logs', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->string('action', 100); // Tên hành động: create_project, resolve_alert, update_member...
            $table->string('entity_type', 50); // Loại thực thể: project, team, alert, user
            $table->unsignedBigInteger('entity_id')->nullable(); // Khóa định danh của thực thể tương ứng
            $table->string('ip_address', 45)->nullable();
            $table->string('user_agent', 255)->nullable();
            $table->jsonb('details')->nullable(); // Dữ liệu chi tiết dạng JSONB tối ưu của PostgreSQL
            $table->timestamp('created_at')->useCurrent();

            // Chỉ mục phục vụ tra cứu theo người dùng và mốc thời gian
            $table->index('user_id');
            $table->index('created_at');
        });
    }

    /**
     * Thu hồi bảng audit_logs khi rollback.
     */
    public function down(): void
    {
        Schema::dropIfExists('audit_logs');
    }
};
