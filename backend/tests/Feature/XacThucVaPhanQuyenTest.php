<?php

namespace Tests\Feature;

use App\Models\NguoiDung;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

/**
 * Kiểm thử toàn diện module Xác thực và Phân quyền (Auth & RBAC - Phase 4).
 * Bao gồm kiểm tra: Đăng ký, Đăng nhập, Profile (Me), Refresh Token, Đăng xuất và Middleware Role.
 */
class XacThucVaPhanQuyenTest extends TestCase
{
    use RefreshDatabase;

    /**
     * Kiểm thử đăng ký tài khoản thành công với thông tin hợp lệ.
     */
    public function test_dang_ky_tai_khoan_thanh_cong(): void
    {
        $duLieuDangKy = [
            'name' => 'Nguyễn Văn Test',
            'email' => 'test.dangky@devops-sentinel.local',
            'password' => 'MatKhau@123456',
            'password_confirmation' => 'MatKhau@123456',
        ];

        $phanHoi = $this->postJson('/api/v1/auth/register', $duLieuDangKy);

        $phanHoi->assertStatus(201)
            ->assertJson([
                'success' => true,
                'data' => [
                    'user' => [
                        'name' => 'Nguyễn Văn Test',
                        'email' => 'test.dangky@devops-sentinel.local',
                        'role' => 'viewer',
                    ],
                    'token_type' => 'Bearer',
                ],
            ]);

        $this->assertDatabaseHas('users', [
            'email' => 'test.dangky@devops-sentinel.local',
            'role' => 'viewer',
        ]);
    }

    /**
     * Kiểm thử đăng ký thất bại khi trùng email.
     */
    public function test_dang_ky_that_bai_khi_trung_email(): void
    {
        NguoiDung::create([
            'name' => 'Người Dùng Cũ',
            'email' => 'da.ton.tai@devops-sentinel.local',
            'password_hash' => password_hash('MatKhau@123', PASSWORD_BCRYPT),
            'role' => 'viewer',
        ]);

        $duLieuTrung = [
            'name' => 'Người Dùng Mới',
            'email' => 'da.ton.tai@devops-sentinel.local',
            'password' => 'MatKhau@123456',
            'password_confirmation' => 'MatKhau@123456',
        ];

        $phanHoi = $this->postJson('/api/v1/auth/register', $duLieuTrung);

        $phanHoi->assertStatus(422)
            ->assertJson([
                'success' => false,
                'error' => [
                    'code' => 'VALIDATION_ERROR',
                ],
            ]);
    }

    /**
     * Kiểm thử đăng nhập thành công cấp phát đúng access token và refresh token.
     */
    public function test_dang_nhap_thanh_cong_cap_token(): void
    {
        $matKhauGoc = 'Sentinel@123456';
        NguoiDung::create([
            'name' => 'Người Dùng Login',
            'email' => 'login.test@devops-sentinel.local',
            'password_hash' => password_hash($matKhauGoc, PASSWORD_BCRYPT),
            'role' => 'viewer',
        ]);

        $phanHoi = $this->postJson('/api/v1/auth/login', [
            'email' => 'login.test@devops-sentinel.local',
            'password' => $matKhauGoc,
        ]);

        $phanHoi->assertStatus(200)
            ->assertJsonStructure([
                'success',
                'data' => [
                    'user' => ['id', 'name', 'email', 'role'],
                    'access_token',
                    'refresh_token',
                    'token_type',
                    'expires_in',
                ],
            ]);
    }

    /**
     * Kiểm thử đăng nhập thất bại khi sai mật khẩu.
     */
    public function test_dang_nhap_that_bai_khi_sai_mat_khau(): void
    {
        NguoiDung::create([
            'name' => 'Người Dùng Test',
            'email' => 'user@devops-sentinel.local',
            'password_hash' => password_hash('DungMatKhau@123', PASSWORD_BCRYPT),
            'role' => 'viewer',
        ]);

        $phanHoi = $this->postJson('/api/v1/auth/login', [
            'email' => 'user@devops-sentinel.local',
            'password' => 'SaiMatKhau@999',
        ]);

        $phanHoi->assertStatus(401)
            ->assertJson([
                'success' => false,
                'error' => [
                    'code' => 'INVALID_CREDENTIALS',
                ],
            ]);
    }

    /**
     * Kiểm thử truy vấn thông tin cá nhân với Bearer token hợp lệ.
     */
    public function test_lay_thong_tin_ca_nhan_qua_token(): void
    {
        $matKhau = 'Sentinel@123';
        $nguoiDung = NguoiDung::create([
            'name' => 'Trần Hồ Sơ',
            'email' => 'hoso@devops-sentinel.local',
            'password_hash' => password_hash($matKhau, PASSWORD_BCRYPT),
            'role' => 'team_lead',
        ]);

        $phanHoiLogin = $this->postJson('/api/v1/auth/login', [
            'email' => 'hoso@devops-sentinel.local',
            'password' => $matKhau,
        ]);

        $chuoiToken = $phanHoiLogin->json('data.access_token');

        $phanHoiMe = $this->withHeader('Authorization', "Bearer {$chuoiToken}")
            ->getJson('/api/v1/auth/me');

        $phanHoiMe->assertStatus(200)
            ->assertJson([
                'success' => true,
                'data' => [
                    'user' => [
                        'name' => 'Trần Hồ Sơ',
                        'email' => 'hoso@devops-sentinel.local',
                        'role' => 'team_lead',
                    ],
                ],
            ]);
    }

    /**
     * Kiểm thử truy cập API được bảo vệ mà không có token bị từ chối 401.
     */
    public function test_truy_cap_khong_token_bi_tu_choi(): void
    {
        $phanHoi = $this->getJson('/api/v1/auth/me');

        $phanHoi->assertStatus(401)
            ->assertJson([
                'success' => false,
                'error' => [
                    'code' => 'UNAUTHORIZED',
                ],
            ]);
    }

    /**
     * Kiểm thử làm mới Access Token bằng Refresh Token hợp lệ.
     */
    public function test_lam_moi_token_thanh_cong(): void
    {
        $matKhau = 'Sentinel@123';
        NguoiDung::create([
            'name' => 'Người Dùng Refresh',
            'email' => 'refresh@devops-sentinel.local',
            'password_hash' => password_hash($matKhau, PASSWORD_BCRYPT),
            'role' => 'viewer',
        ]);

        $phanHoiLogin = $this->postJson('/api/v1/auth/login', [
            'email' => 'refresh@devops-sentinel.local',
            'password' => $matKhau,
        ]);

        $chuoiRefreshToken = $phanHoiLogin->json('data.refresh_token');

        $phanHoiRefresh = $this->postJson('/api/v1/auth/refresh', [
            'refresh_token' => $chuoiRefreshToken,
        ]);

        $phanHoiRefresh->assertStatus(200)
            ->assertJsonStructure([
                'success',
                'data' => [
                    'access_token',
                    'refresh_token',
                    'token_type',
                    'expires_in',
                ],
            ]);
    }

    /**
     * Kiểm thử phân quyền RBAC: Quản trị viên truy cập được vùng Admin, Viewer bị chặn 403.
     */
    public function test_phan_quyen_vai_tro_admin(): void
    {
        $matKhau = 'Sentinel@123';

        // 1. Tạo Admin
        NguoiDung::create([
            'name' => 'Sếp Quản Trị',
            'email' => 'boss@devops-sentinel.local',
            'password_hash' => password_hash($matKhau, PASSWORD_BCRYPT),
            'role' => 'admin',
        ]);

        // 2. Tạo Viewer
        NguoiDung::create([
            'name' => 'Nhân Viên Xem',
            'email' => 'nhanvien@devops-sentinel.local',
            'password_hash' => password_hash($matKhau, PASSWORD_BCRYPT),
            'role' => 'viewer',
        ]);

        // Đăng nhập lấy token Admin
        $tokenAdmin = $this->postJson('/api/v1/auth/login', [
            'email' => 'boss@devops-sentinel.local',
            'password' => $matKhau,
        ])->json('data.access_token');

        // Đăng nhập lấy token Viewer
        $tokenViewer = $this->postJson('/api/v1/auth/login', [
            'email' => 'nhanvien@devops-sentinel.local',
            'password' => $matKhau,
        ])->json('data.access_token');

        // Admin truy cập tuyến admin -> 200 OK
        $phanHoiAdmin = $this->withHeader('Authorization', "Bearer {$tokenAdmin}")
            ->getJson('/api/v1/admin/kiem-tra-quyen');
        $phanHoiAdmin->assertStatus(200);

        // Viewer truy cập tuyến admin -> 403 FORBIDDEN
        $phanHoiViewer = $this->withHeader('Authorization', "Bearer {$tokenViewer}")
            ->getJson('/api/v1/admin/kiem-tra-quyen');
        $phanHoiViewer->assertStatus(403)
            ->assertJson([
                'success' => false,
                'error' => [
                    'code' => 'FORBIDDEN',
                ],
            ]);
    }

    /**
     * Kiểm thử đăng xuất thành công và vô hiệu hóa refresh token.
     */
    public function test_dang_xuat_thu_hoi_token(): void
    {
        $matKhau = 'Sentinel@123';
        NguoiDung::create([
            'name' => 'Người Dùng Logout',
            'email' => 'logout@devops-sentinel.local',
            'password_hash' => password_hash($matKhau, PASSWORD_BCRYPT),
            'role' => 'viewer',
        ]);

        $phanHoiLogin = $this->postJson('/api/v1/auth/login', [
            'email' => 'logout@devops-sentinel.local',
            'password' => $matKhau,
        ]);

        $tokenAccess = $phanHoiLogin->json('data.access_token');
        $tokenRefresh = $phanHoiLogin->json('data.refresh_token');

        // Thực hiện đăng xuất
        $phanHoiLogout = $this->withHeader('Authorization', "Bearer {$tokenAccess}")
            ->postJson('/api/v1/auth/logout');

        $phanHoiLogout->assertStatus(200)
            ->assertJson([
                'success' => true,
                'message' => 'Đăng xuất tài khoản thành công.',
            ]);

        // Dùng lại refresh token cũ để làm mới phiên -> Phải bị từ chối 401
        $phanHoiRefreshLai = $this->postJson('/api/v1/auth/refresh', [
            'refresh_token' => $tokenRefresh,
        ]);

        $phanHoiRefreshLai->assertStatus(401);
    }
}
