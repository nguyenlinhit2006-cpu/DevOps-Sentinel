<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Khởi tạo bảng dự án (projects).
     * Mục đích: Quản lý thông tin từng repository, cấu hình nhà cung cấp CI/CD và secret đối soát chữ ký HMAC.
     */
    public function up(): void
    {
        Schema::create('projects', function (Blueprint $table) {
            $table->id();
            $table->foreignId('team_id')->nullable()->constrained('teams')->nullOnDelete();
            $table->string('name', 100);
            $table->string('slug', 120)->unique();
            $table->text('description')->nullable();
            $table->string('repository_url', 255)->nullable();
            $table->string('ci_provider', 50)->default('github'); // Nhà cung cấp CI: github, gitlab, generic
            $table->string('webhook_secret', 255); // Khóa bí mật dùng để kiểm tra chữ ký HMAC SHA256
            $table->integer('failure_threshold')->default(3); // Số lần thất bại liên tiếp để kích hoạt Sentinel Alert
            $table->foreignId('created_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamps();

            // Chỉ mục tối ưu hóa truy vấn danh sách dự án theo nhóm
            $table->index('team_id');
        });
    }

    /**
     * Thu hồi bảng dự án khi rollback.
     */
    public function down(): void
    {
        Schema::dropIfExists('projects');
    }
};
