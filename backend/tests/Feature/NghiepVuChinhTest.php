<?php

namespace Tests\Feature;

use App\Models\CanhBao;
use App\Models\DuAn;
use App\Models\LuotChayPipeline;
use App\Models\NguoiDung;
use App\Models\Nhom;
use App\Services\DichVuJwt;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

/**
 * Kiểm thử toàn diện các nghiệp vụ cốt lõi Backend (Phase 5).
 * Bao gồm: CRUD Dự án & Nhóm, Webhook Ingestion HMAC, Logic Cảnh báo Sentinel (fail N lần), Dashboard & Audit Log.
 */
class NghiepVuChinhTest extends TestCase
{
    use RefreshDatabase;

    protected DichVuJwt $dichVuJwt;
    protected NguoiDung $quanTriVien;
    protected NguoiDung $truongNhom;
    protected NguoiDung $nguoiXem;
    protected string $tokenAdmin;
    protected string $tokenLead;
    protected string $tokenViewer;

    protected function setUp(): void
    {
        parent::setUp();
        $this->dichVuJwt = app(DichVuJwt::class);

        $matKhau = password_hash('Sentinel@123', PASSWORD_BCRYPT);

        $this->quanTriVien = NguoiDung::create([
            'name' => 'Quản Trị Viên',
            'email' => 'admin@test.local',
            'password_hash' => $matKhau,
            'role' => 'admin',
        ]);

        $this->truongNhom = NguoiDung::create([
            'name' => 'Trưởng Nhóm',
            'email' => 'lead@test.local',
            'password_hash' => $matKhau,
            'role' => 'team_lead',
        ]);

        $this->nguoiXem = NguoiDung::create([
            'name' => 'Người Xem',
            'email' => 'viewer@test.local',
            'password_hash' => $matKhau,
            'role' => 'viewer',
        ]);

        $this->tokenAdmin = $this->dichVuJwt->taoAccessToken($this->quanTriVien);
        $this->tokenLead = $this->dichVuJwt->taoAccessToken($this->truongNhom);
        $this->tokenViewer = $this->dichVuJwt->taoAccessToken($this->nguoiXem);
    }

    /**
     * Kiểm thử CRUD Nhóm làm việc và phân bổ thành viên.
     */
    public function test_crud_nhom_va_thanh_vien(): void
    {
        // 1. Quản trị viên tạo nhóm mới -> 201 Created
        $phanHoiTao = $this->withHeader('Authorization', "Bearer {$this->tokenAdmin}")
            ->postJson('/api/v1/teams', [
                'name' => 'Nhóm Core Backend',
                'description' => 'Chuyên phát triển API và Cơ sở dữ liệu',
            ]);
        $phanHoiTao->assertStatus(201);
        $idNhom = $phanHoiTao->json('data.id');

        // 2. Người xem thử tạo nhóm -> 403 Forbidden
        $this->withHeader('Authorization', "Bearer {$this->tokenViewer}")
            ->postJson('/api/v1/teams', ['name' => 'Nhóm Trái Phép'])
            ->assertStatus(403);

        // 3. Thêm thành viên vào nhóm
        $phanHoiThemTv = $this->withHeader('Authorization', "Bearer {$this->tokenAdmin}")
            ->postJson("/api/v1/teams/{$idNhom}/members", [
                'user_id' => $this->truongNhom->id,
                'role' => 'team_lead',
            ]);
        $phanHoiThemTv->assertStatus(200);

        // 4. Xem chi tiết nhóm
        $phanHoiChiTiet = $this->withHeader('Authorization', "Bearer {$this->tokenViewer}")
            ->getJson("/api/v1/teams/{$idNhom}");
        $phanHoiChiTiet->assertStatus(200)
            ->assertJsonPath('data.name', 'Nhóm Core Backend');
    }

    /**
     * Kiểm thử CRUD Dự án và sinh lại Webhook Secret.
     */
    public function test_crud_du_an(): void
    {
        // 1. Trưởng nhóm tạo dự án mới -> 201 Created kèm webhook_secret
        $phanHoiTao = $this->withHeader('Authorization', "Bearer {$this->tokenLead}")
            ->postJson('/api/v1/projects', [
                'name' => 'Hệ thống Thanh toán V2',
                'description' => 'Cổng kết nối ngân hàng',
                'repository_url' => 'https://github.com/company/payment-v2',
                'ci_provider' => 'github',
                'failure_threshold' => 3,
            ]);

        $phanHoiTao->assertStatus(201)
            ->assertJsonStructure([
                'success',
                'data' => [
                    'id', 'name', 'slug', 'webhook_secret', 'webhook_url',
                ],
            ]);

        $idDuAn = $phanHoiTao->json('data.id');
        $secretCu = $phanHoiTao->json('data.webhook_secret');

        // 2. Người xem thử tạo dự án -> 403 Forbidden
        $this->withHeader('Authorization', "Bearer {$this->tokenViewer}")
            ->postJson('/api/v1/projects', ['name' => 'Dự án Không Quyền'])
            ->assertStatus(403);

        // 3. Cập nhật dự án
        $phanHoiCapNhat = $this->withHeader('Authorization', "Bearer {$this->tokenLead}")
            ->putJson("/api/v1/projects/{$idDuAn}", [
                'description' => 'Mô tả đã cập nhật',
                'failure_threshold' => 5,
            ]);
        $phanHoiCapNhat->assertStatus(200)
            ->assertJsonPath('data.failure_threshold', 5);

        // 4. Sinh lại Webhook Secret
        $phanHoiSecret = $this->withHeader('Authorization', "Bearer {$this->tokenLead}")
            ->postJson("/api/v1/projects/{$idDuAn}/regenerate-secret");
        $phanHoiSecret->assertStatus(200);
        $secretMoi = $phanHoiSecret->json('data.webhook_secret');
        $this->assertNotEquals($secretCu, $secretMoi);

        // 5. Quản trị viên xóa dự án -> 200 OK
        $this->withHeader('Authorization', "Bearer {$this->tokenAdmin}")
            ->deleteJson("/api/v1/projects/{$idDuAn}")
            ->assertStatus(200);
    }

    /**
     * Kiểm thử tiếp nhận Webhook với xác thực chữ ký HMAC SHA-256 (GitHub & Generic) và Secret Token (GitLab).
     */
    public function test_webhook_ingestion_va_xac_thuc_hmac(): void
    {
        $khoaBiMat = 'secret_webhook_key_12345';
        $duAn = DuAn::create([
            'name' => 'Dự Án Test Webhook',
            'slug' => 'du-an-test-webhook',
            'ci_provider' => 'github',
            'webhook_secret' => $khoaBiMat,
            'failure_threshold' => 3,
        ]);

        $duLieuPayload = [
            'pipeline_name' => 'unit-test',
            'commit_hash' => 'abc123def456',
            'commit_message' => 'feat: thêm webhook controller',
            'branch' => 'main',
            'author' => 'developer',
            'status' => 'success',
            'duration_seconds' => 35,
            'short_log' => 'Tất cả 10 kiểm thử đều thành công.',
        ];
        $chuoiJson = json_encode($duLieuPayload);

        // 1. Gửi webhook không kèm chữ ký -> 401 Unauthorized
        $this->call('POST', "/api/v1/webhooks/projects/{$duAn->id}", [], [], [], [], $chuoiJson)
            ->assertStatus(401)
            ->assertJsonPath('error.code', 'INVALID_SIGNATURE');

        // 2. Gửi webhook sai chữ ký -> 401 Unauthorized
        $this->call('POST', "/api/v1/webhooks/projects/{$duAn->id}", [], [], [], [
            'HTTP_X_Hub_Signature_256' => 'sha256=chu_ky_sai_hoan_toan',
            'CONTENT_TYPE' => 'application/json',
        ], $chuoiJson)
            ->assertStatus(401);

        // 3. Gửi webhook đúng chữ ký GitHub HMAC-SHA256 -> 200 OK
        $chuKyDung = 'sha256=' . hash_hmac('sha256', $chuoiJson, $khoaBiMat);
        $phanHoiDung = $this->call('POST', "/api/v1/webhooks/projects/{$duAn->id}", [], [], [], [
            'HTTP_X_Hub_Signature_256' => $chuKyDung,
            'CONTENT_TYPE' => 'application/json',
        ], $chuoiJson);

        $phanHoiDung->assertStatus(200)
            ->assertJsonPath('success', true)
            ->assertJsonPath('data.status', 'processed');

        $this->assertDatabaseHas('pipeline_runs', [
            'project_id' => $duAn->id,
            'commit_hash' => 'abc123def456',
            'status' => 'success',
        ]);

        // 4. Gửi webhook theo chuẩn GitLab Token -> 200 OK
        $duLieuGitlab = array_merge($duLieuPayload, ['commit_hash' => 'gitlab_hash_789']);
        $chuoiJsonGitlab = json_encode($duLieuGitlab);

        $phanHoiGitlab = $this->call('POST', "/api/v1/webhooks/projects/{$duAn->id}", [], [], [], [
            'HTTP_X_Gitlab_Token' => $khoaBiMat,
            'CONTENT_TYPE' => 'application/json',
        ], $chuoiJsonGitlab);

        $phanHoiGitlab->assertStatus(200);
        $this->assertDatabaseHas('pipeline_runs', [
            'project_id' => $duAn->id,
            'commit_hash' => 'gitlab_hash_789',
        ]);
    }

    /**
     * Kiểm thử logic Sentinel Alert: Kích hoạt khi pipeline thất bại liên tiếp N lần và tự phục hồi khi thành công.
     */
    public function test_logic_canh_bao_that_bai_lien_tiep_n_lan(): void
    {
        $khoaBiMat = 'secret_test_alert';
        $nguongLoi = 3;

        $duAn = DuAn::create([
            'name' => 'Dự Án Giám Sát Alert',
            'slug' => 'du-an-giam-sat-alert',
            'ci_provider' => 'generic',
            'webhook_secret' => $khoaBiMat,
            'failure_threshold' => $nguongLoi,
        ]);

        $guiWebhook = function (string $trangThai, string $commit) use ($duAn, $khoaBiMat) {
            $payload = [
                'pipeline_name' => 'ci-test',
                'commit_hash' => $commit,
                'branch' => 'main',
                'status' => $trangThai,
                'short_log' => "Chi tiết lỗi của commit {$commit}",
            ];
            $json = json_encode($payload);
            $signature = 'sha256=' . hash_hmac('sha256', $json, $khoaBiMat);

            return $this->call('POST', "/api/v1/webhooks/projects/{$duAn->id}", [], [], [], [
                'HTTP_X_Sentinel_Signature' => $signature,
                'CONTENT_TYPE' => 'application/json',
            ], $json);
        };

        // Lần 1: Thất bại (Lỗi 1/3) -> Chưa bật cảnh báo
        $phanHoi1 = $guiWebhook('failed', 'commit_fail_1');
        $phanHoi1->assertStatus(200)->assertJsonPath('data.alert_triggered', false);
        $this->assertDatabaseCount('alerts', 0);

        // Lần 2: Thất bại (Lỗi 2/3) -> Chưa bật cảnh báo
        $phanHoi2 = $guiWebhook('failed', 'commit_fail_2');
        $phanHoi2->assertStatus(200)->assertJsonPath('data.alert_triggered', false);
        $this->assertDatabaseCount('alerts', 0);

        // Lần 3: Thất bại (Lỗi 3/3 = Đạt ngưỡng) -> KÍCH HOẠT CẢNH BÁO
        $phanHoi3 = $guiWebhook('failed', 'commit_fail_3');
        $phanHoi3->assertStatus(200)->assertJsonPath('data.alert_triggered', true);

        $this->assertDatabaseHas('alerts', [
            'project_id' => $duAn->id,
            'status' => 'active',
            'consecutive_failures' => 3,
        ]);

        // Lần 4: Tiếp tục thất bại -> Cập nhật cảnh báo lên 4 lần
        $guiWebhook('failed', 'commit_fail_4');
        $this->assertDatabaseHas('alerts', [
            'project_id' => $duAn->id,
            'status' => 'active',
            'consecutive_failures' => 4,
        ]);

        // Lần 5: Thành công (Success) -> Cảnh báo tự động phục hồi (resolved)
        $guiWebhook('success', 'commit_success_5');
        $this->assertDatabaseHas('alerts', [
            'project_id' => $duAn->id,
            'status' => 'resolved',
        ]);
    }

    /**
     * Kiểm thử người dùng đóng cảnh báo thủ công.
     */
    public function test_dong_canh_bao_thu_cong(): void
    {
        $duAn = DuAn::create([
            'name' => 'Dự án Cảnh Báo',
            'slug' => 'du-an-canh-bao',
            'webhook_secret' => 'secret_123',
        ]);

        $canhBao = CanhBao::create([
            'project_id' => $duAn->id,
            'title' => 'Cảnh báo mẫu',
            'message' => 'Lỗi phát sinh',
            'status' => 'active',
            'consecutive_failures' => 3,
        ]);

        // Trưởng nhóm đóng cảnh báo -> 200 OK
        $phanHoi = $this->withHeader('Authorization', "Bearer {$this->tokenLead}")
            ->postJson("/api/v1/alerts/{$canhBao->id}/resolve");

        $phanHoi->assertStatus(200)
            ->assertJsonPath('data.status', 'resolved');

        $this->assertDatabaseHas('alerts', [
            'id' => $canhBao->id,
            'status' => 'resolved',
            'resolved_by' => $this->truongNhom->id,
        ]);
    }

    /**
     * Kiểm thử các API Dashboard tổng quan và thống kê theo ngày (Aggregate Query).
     */
    public function test_api_dashboard_va_so_lieu_thong_ke(): void
    {
        $duAn = DuAn::create([
            'name' => 'Dự án Metrics',
            'slug' => 'du-an-metrics',
            'webhook_secret' => 'secret_metric',
        ]);

        LuotChayPipeline::create([
            'project_id' => $duAn->id,
            'branch' => 'main',
            'status' => 'success',
            'commit_hash' => 'hash_1',
            'created_at' => now(),
        ]);

        LuotChayPipeline::create([
            'project_id' => $duAn->id,
            'branch' => 'main',
            'status' => 'failed',
            'commit_hash' => 'hash_2',
            'created_at' => now(),
        ]);

        // 1. Kiểm tra Dashboard Summary
        $phanHoiSummary = $this->withHeader('Authorization', "Bearer {$this->tokenViewer}")
            ->getJson('/api/v1/dashboard/summary');

        $phanHoiSummary->assertStatus(200)
            ->assertJsonPath('success', true)
            ->assertJsonPath('data.total_projects', 1)
            ->assertJsonPath('data.total_runs', 2);
        $this->assertEquals(50, (float) $phanHoiSummary->json('data.success_rate_percent'));

        // 2. Kiểm tra Dashboard Metrics chuỗi ngày
        $phanHoiMetrics = $this->withHeader('Authorization', "Bearer {$this->tokenViewer}")
            ->getJson('/api/v1/dashboard/metrics?days=7');

        $phanHoiMetrics->assertStatus(200)
            ->assertJsonPath('data.days', 7)
            ->assertJsonCount(7, 'data.daily_metrics');
    }

    /**
     * Kiểm thử truy xuất Audit Log và đảm bảo chỉ Quản trị viên mới được phép truy cập.
     */
    public function test_truy_van_audit_log_phan_quyen(): void
    {
        // 1. Quản trị viên truy vấn danh sách log -> 200 OK
        $phanHoiAdmin = $this->withHeader('Authorization', "Bearer {$this->tokenAdmin}")
            ->getJson('/api/v1/audit-logs');
        $phanHoiAdmin->assertStatus(200);

        // 2. Trưởng nhóm hoặc Người xem truy vấn -> 403 Forbidden
        $this->withHeader('Authorization', "Bearer {$this->tokenLead}")
            ->getJson('/api/v1/audit-logs')
            ->assertStatus(403);
    }
}
