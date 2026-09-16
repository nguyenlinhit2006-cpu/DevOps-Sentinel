<?php

namespace Database\Seeders;

use App\Models\CanhBao;
use App\Models\DuAn;
use App\Models\LuotChayPipeline;
use App\Models\NguoiDung;
use App\Models\NhatKyHoatDong;
use App\Models\Nhom;
use Illuminate\Database\Seeder;
use Illuminate\Support\Carbon;

/**
 * Seeder khởi tạo dữ liệu mẫu cho hệ thống DevOps Sentinel.
 * Cung cấp tài khoản quản trị, trưởng nhóm, người xem, dự án mẫu và lịch sử pipeline để kiểm thử.
 */
class DatabaseSeeder extends Seeder
{
    /**
     * Thực thi nạp dữ liệu mẫu vào cơ sở dữ liệu PostgreSQL.
     */
    public function run(): void
    {
        $danhSachTaiKhoan = $this->khoiTaoNguoiDung();
        $nhomLamViec = $this->khoiTaoNhomVaThanhVien($danhSachTaiKhoan);
        $danhSachDuAn = $this->khoiTaoDuAn($danhSachTaiKhoan, $nhomLamViec);
        $this->khoiTaoLuotChayVaCanhBao($danhSachDuAn);
        $this->khoiTaoNhatKyHeThong($danhSachTaiKhoan['quanTriVien']);
    }

    /**
     * Khởi tạo các tài khoản người dùng mẫu tương ứng 3 vai trò (Admin, Team Lead, Viewer).
     */
    private function khoiTaoNguoiDung(): array
    {
        $matKhauChung = password_hash('Sentinel@123456', PASSWORD_BCRYPT);

        $quanTriVien = NguoiDung::create([
            'name' => 'Nguyễn Quản Trị',
            'email' => 'admin@devops-sentinel.local',
            'password_hash' => $matKhauChung,
            'role' => 'admin',
        ]);

        $truongNhom = NguoiDung::create([
            'name' => 'Trần Trưởng Nhóm',
            'email' => 'lead@devops-sentinel.local',
            'password_hash' => $matKhauChung,
            'role' => 'team_lead',
        ]);

        $nguoiXem = NguoiDung::create([
            'name' => 'Lê Người Xem',
            'email' => 'viewer@devops-sentinel.local',
            'password_hash' => $matKhauChung,
            'role' => 'viewer',
        ]);

        return [
            'quanTriVien' => $quanTriVien,
            'truongNhom' => $truongNhom,
            'nguoiXem' => $nguoiXem,
        ];
    }

    /**
     * Khởi tạo nhóm làm việc mẫu và gán thành viên vào nhóm.
     */
    private function khoiTaoNhomVaThanhVien(array $danhSachTaiKhoan): Nhom
    {
        $nhomDevOps = Nhom::create([
            'name' => 'Nhóm Hạ Tầng & SRE',
            'description' => 'Chịu trách nhiệm giám sát độ sẵn sàng và triển khai hệ thống toàn công ty',
            'created_by' => $danhSachTaiKhoan['quanTriVien']->id,
        ]);

        // Gán trưởng nhóm và người xem vào nhóm
        $nhomDevOps->danhSachThanhVien()->attach($danhSachTaiKhoan['truongNhom']->id, ['role' => 'team_lead']);
        $nhomDevOps->danhSachThanhVien()->attach($danhSachTaiKhoan['nguoiXem']->id, ['role' => 'viewer']);

        return $nhomDevOps;
    }

    /**
     * Khởi tạo các dự án mẫu có cấu hình Webhook và ngưỡng cảnh báo.
     */
    private function khoiTaoDuAn(array $danhSachTaiKhoan, Nhom $nhomLamViec): array
    {
        $duAnPortal = DuAn::create([
            'team_id' => $nhomLamViec->id,
            'name' => 'Frontend Portal WASM',
            'slug' => 'frontend-portal-wasm',
            'description' => 'Giao diện giám sát tập trung viết bằng Go WebAssembly',
            'repository_url' => 'https://github.com/nguyenlinhit2006-cpu/DevOps-Sentinel',
            'ci_provider' => 'github',
            'webhook_secret' => 'sec_frontend_wasm_9a8b7c6d5e4f3a2b1c0d',
            'failure_threshold' => 3,
            'created_by' => $danhSachTaiKhoan['truongNhom']->id,
        ]);

        $duAnBackend = DuAn::create([
            'team_id' => $nhomLamViec->id,
            'name' => 'Backend API Core',
            'slug' => 'backend-api-core',
            'description' => 'Hệ thống RESTful API lõi xử lý dữ liệu và tiếp nhận webhook',
            'repository_url' => 'https://gitlab.com/company/backend-api-core',
            'ci_provider' => 'gitlab',
            'webhook_secret' => 'sec_backend_api_1f2e3d4c5b6a7b8c9d0e',
            'failure_threshold' => 3,
            'created_by' => $danhSachTaiKhoan['quanTriVien']->id,
        ]);

        return [
            'duAnPortal' => $duAnPortal,
            'duAnBackend' => $duAnBackend,
        ];
    }

    /**
     * Khởi tạo lịch sử chạy pipeline và mô phỏng tình huống cảnh báo khi fail liên tiếp 3 lần.
     */
    private function khoiTaoLuotChayVaCanhBao(array $danhSachDuAn): void
    {
        $duAnPortal = $danhSachDuAn['duAnPortal'];
        $duAnBackend = $danhSachDuAn['duAnBackend'];
        $thoiDiemHienTai = Carbon::now();

        // 1. Dự án Portal WASM: Chạy thành công ổn định
        LuotChayPipeline::create([
            'project_id' => $duAnPortal->id,
            'pipeline_name' => 'build-and-test',
            'commit_hash' => 'e7b1a2c3d4f5',
            'commit_message' => 'feat: tối ưu hóa bộ nhớ WASM DOM wrapper',
            'branch' => 'main',
            'author' => 'linh.nguyen',
            'status' => 'success',
            'trigger_event' => 'push',
            'duration_seconds' => 48,
            'short_log' => 'Build succeeded. All 18 unit tests passed in 0.42s.',
            'external_url' => 'https://github.com/nguyenlinhit2006-cpu/DevOps-Sentinel/actions/runs/101',
            'started_at' => $thoiDiemHienTai->copy()->subMinutes(60),
            'finished_at' => $thoiDiemHienTai->copy()->subMinutes(59),
            'created_at' => $thoiDiemHienTai->copy()->subMinutes(60),
        ]);

        LuotChayPipeline::create([
            'project_id' => $duAnPortal->id,
            'pipeline_name' => 'build-and-test',
            'commit_hash' => 'c4d5e6f7a8b9',
            'commit_message' => 'fix: điều chỉnh kích thước container canvas',
            'branch' => 'main',
            'author' => 'linh.nguyen',
            'status' => 'success',
            'trigger_event' => 'push',
            'duration_seconds' => 45,
            'short_log' => 'Build succeeded. All 18 unit tests passed in 0.39s.',
            'external_url' => 'https://github.com/nguyenlinhit2006-cpu/DevOps-Sentinel/actions/runs/102',
            'started_at' => $thoiDiemHienTai->copy()->subMinutes(25),
            'finished_at' => $thoiDiemHienTai->copy()->subMinutes(24),
            'created_at' => $thoiDiemHienTai->copy()->subMinutes(25),
        ]);

        // 2. Dự án Backend Core: Mô phỏng 3 lần thất bại liên tiếp dẫn tới cảnh báo Sentinel
        LuotChayPipeline::create([
            'project_id' => $duAnBackend->id,
            'pipeline_name' => 'ci-cd-pipeline',
            'commit_hash' => '8a9b0c1d2e3f',
            'commit_message' => 'test: cập nhật kiểm thử tích hợp cơ sở dữ liệu',
            'branch' => 'main',
            'author' => 'dev.lead',
            'status' => 'failed',
            'trigger_event' => 'push',
            'duration_seconds' => 115,
            'short_log' => "Lỗi kiểm thử: Database connection timeout on port 5432\nStack trace: PDOException in Connection.php line 42",
            'external_url' => 'https://gitlab.com/company/backend-api-core/-/pipelines/201',
            'started_at' => $thoiDiemHienTai->copy()->subMinutes(40),
            'finished_at' => $thoiDiemHienTai->copy()->subMinutes(38),
            'created_at' => $thoiDiemHienTai->copy()->subMinutes(40),
        ]);

        LuotChayPipeline::create([
            'project_id' => $duAnBackend->id,
            'pipeline_name' => 'ci-cd-pipeline',
            'commit_hash' => '9b0c1d2e3f4a',
            'commit_message' => 'fix: tăng thời gian timeout kết nối cơ sở dữ liệu',
            'branch' => 'main',
            'author' => 'dev.lead',
            'status' => 'failed',
            'trigger_event' => 'push',
            'duration_seconds' => 120,
            'short_log' => "Lỗi kiểm thử: Migration 2026_create_users failed: relation already exists\nSQLSTATE[42P07]",
            'external_url' => 'https://gitlab.com/company/backend-api-core/-/pipelines/202',
            'started_at' => $thoiDiemHienTai->copy()->subMinutes(20),
            'finished_at' => $thoiDiemHienTai->copy()->subMinutes(18),
            'created_at' => $thoiDiemHienTai->copy()->subMinutes(20),
        ]);

        $luotChayLoiCuoi = LuotChayPipeline::create([
            'project_id' => $duAnBackend->id,
            'pipeline_name' => 'ci-cd-pipeline',
            'commit_hash' => '0c1d2e3f4a5b',
            'commit_message' => 'fix: sửa cú pháp rollback migration',
            'branch' => 'main',
            'author' => 'dev.lead',
            'status' => 'failed',
            'trigger_event' => 'push',
            'duration_seconds' => 110,
            'short_log' => "Lỗi kiểm thử: Schema table locks failed\nExit code: 1",
            'external_url' => 'https://gitlab.com/company/backend-api-core/-/pipelines/203',
            'started_at' => $thoiDiemHienTai->copy()->subMinutes(5),
            'finished_at' => $thoiDiemHienTai->copy()->subMinutes(3),
            'created_at' => $thoiDiemHienTai->copy()->subMinutes(5),
        ]);

        // Tạo cảnh báo tự động vì đạt ngưỡng 3 lần lỗi liên tiếp
        CanhBao::create([
            'project_id' => $duAnBackend->id,
            'last_pipeline_run_id' => $luotChayLoiCuoi->id,
            'title' => 'Pipeline ci-cd-pipeline thất bại 3 lần liên tiếp trên nhánh main',
            'message' => 'Dự án Backend API Core gặp sự cố nghiêm trọng: đã có 3 lượt build thất bại liên tục. Lỗi gần nhất: Schema table locks failed (Exit code 1).',
            'consecutive_failures' => 3,
            'status' => 'active',
        ]);
    }

    /**
     * Khởi tạo bản ghi kiểm toán mẫu đầu tiên của hệ thống.
     */
    private function khoiTaoNhatKyHeThong(NguoiDung $quanTriVien): void
    {
        NhatKyHoatDong::create([
            'user_id' => $quanTriVien->id,
            'action' => 'khoi_tao_he_thong',
            'entity_type' => 'system',
            'entity_id' => null,
            'ip_address' => '127.0.0.1',
            'user_agent' => 'DevOps-Sentinel-Seeder/1.0',
            'details' => [
                'thong_diep' => 'Khởi tạo dữ liệu mẫu kiểm thử ban đầu thành công',
                'phien_ban' => '1.0.0-phase3',
            ],
            'created_at' => Carbon::now(),
        ]);
    }
}
