<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Khởi tạo bảng nhóm (teams) và phân bổ thành viên nhóm (team_members).
     * Mục đích: Cho phép tổ chức các dự án theo từng nhóm làm việc và phân quyền cụ thể.
     */
    public function up(): void
    {
        // Bảng danh sách các nhóm phát triển
        Schema::create('teams', function (Blueprint $table) {
            $table->id();
            $table->string('name', 100);
            $table->text('description')->nullable();
            $table->foreignId('created_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamps();
        });

        // Bảng liên kết thành viên và nhóm kèm vai trò nội bộ
        Schema::create('team_members', function (Blueprint $table) {
            $table->id();
            $table->foreignId('team_id')->constrained('teams')->cascadeOnDelete();
            $table->foreignId('user_id')->constrained('users')->cascadeOnDelete();
            $table->string('role', 20)->default('viewer'); // Vai trò trong nhóm: team_lead, viewer
            $table->timestamp('created_at')->useCurrent();

            // Ràng buộc duy nhất: Một người dùng chỉ tham gia một nhóm với một vai trò xác định
            $table->unique(['team_id', 'user_id']);
        });
    }

    /**
     * Thu hồi bảng nhóm và thành viên nhóm khi rollback.
     */
    public function down(): void
    {
        Schema::dropIfExists('team_members');
        Schema::dropIfExists('teams');
    }
};
