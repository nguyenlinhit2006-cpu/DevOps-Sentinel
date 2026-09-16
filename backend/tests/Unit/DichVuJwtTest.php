<?php

namespace Tests\Unit;

use App\Models\NguoiDung;
use App\Services\DichVuJwt;
use Tests\TestCase;

/**
 * Kiểm thử đơn vị (Unit Test) cho dịch vụ DichVuJwt.
 * Mục đích: Đảm bảo tính toàn vẹn và bảo mật của thuật toán ký HMAC-SHA256.
 */
class DichVuJwtTest extends TestCase
{
    protected DichVuJwt $dichVuJwt;

    protected function setUp(): void
    {
        parent::setUp();
        $this->dichVuJwt = new DichVuJwt();
    }

    /**
     * Kiểm thử tạo và giải mã Access Token hợp lệ.
     */
    public function test_tao_va_xac_thuc_token_hop_le(): void
    {
        $nguoiDungGiaLap = new NguoiDung([
            'name' => 'Nguyễn Văn Test',
            'email' => 'jwt.unit@test.local',
            'role' => 'admin',
        ]);
        $nguoiDungGiaLap->id = 999;

        $chuoiToken = $this->dichVuJwt->taoAccessToken($nguoiDungGiaLap);
        $this->assertNotEmpty($chuoiToken);

        $duLieuGiaiMa = $this->dichVuJwt->xacThucAccessToken($chuoiToken);

        $this->assertNotNull($duLieuGiaiMa);
        $this->assertEquals(999, $duLieuGiaiMa->sub);
        $this->assertEquals('jwt.unit@test.local', $duLieuGiaiMa->email);
        $this->assertEquals('admin', $duLieuGiaiMa->role);
    }

    /**
     * Kiểm thử phát hiện token bị giả mạo chữ ký.
     */
    public function test_phat_hien_token_gia_mao(): void
    {
        $nguoiDungGiaLap = new NguoiDung([
            'name' => 'Nguyễn Văn Test',
            'email' => 'fake@test.local',
            'role' => 'viewer',
        ]);
        $nguoiDungGiaLap->id = 100;

        $chuoiToken = $this->dichVuJwt->taoAccessToken($nguoiDungGiaLap);

        // Giả mạo payload hoặc chữ ký
        $tokenGiaMao = $chuoiToken . 'tampered';
        $duLieuGiaiMa = $this->dichVuJwt->xacThucAccessToken($tokenGiaMao);

        $this->assertNull($duLieuGiaiMa);
    }

    /**
     * Kiểm thử tạo chuỗi Refresh Token ngẫu nhiên và băm SHA-256 nhất quán.
     */
    public function test_tao_va_bam_refresh_token(): void
    {
        $refreshToken1 = $this->dichVuJwt->taoRefreshToken();
        $refreshToken2 = $this->dichVuJwt->taoRefreshToken();

        $this->assertEquals(64, strlen($refreshToken1));
        $this->assertNotEquals($refreshToken1, $refreshToken2);

        $banBam1 = $this->dichVuJwt->bamRefreshToken($refreshToken1);
        $banBam2 = $this->dichVuJwt->bamRefreshToken($refreshToken1);

        $this->assertEquals($banBam1, $banBam2);
        $this->assertEquals(64, strlen($banBam1));
    }
}
