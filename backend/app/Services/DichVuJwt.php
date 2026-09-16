<?php

namespace App\Services;

use App\Models\NguoiDung;
use Exception;

/**
 * Lớp dịch vụ xử lý cấp phát, mã hóa và giải mã JSON Web Token (JWT).
 * Sử dụng thuật toán ký HMAC-SHA256 (HS256) bảo mật, độc lập không phụ thuộc thư viện ngoài.
 */
class DichVuJwt
{
    private string $khoaBiMat;
    private int $thoiGianSongAccess;

    public function __construct()
    {
        $this->khoaBiMat = env('JWT_SECRET', 'devops_sentinel_super_secret_jwt_key_32bytes_min_2026');
        $this->thoiGianSongAccess = (int) env('JWT_TTL', 900); // Mặc định 15 phút (900 giây)
    }

    /**
     * Tạo Access Token mới cho người dùng.
     */
    public function taoAccessToken(NguoiDung $nguoiDung): string
    {
        $thoiDiemHienTai = time();
        $thoiDiemHetHan = $thoiDiemHienTai + $this->thoiGianSongAccess;

        $tieuDe = [
            'typ' => 'JWT',
            'alg' => 'HS256',
        ];

        $duLieuTai = [
            'iss' => 'devops-sentinel-api',
            'sub' => $nguoiDung->id,
            'email' => $nguoiDung->email,
            'role' => $nguoiDung->role,
            'iat' => $thoiDiemHienTai,
            'exp' => $thoiDiemHetHan,
        ];

        $chuoiTieuDeBase64 = $this->maHoaBase64Url(json_encode($tieuDe));
        $chuoiDuLieuBase64 = $this->maHoaBase64Url(json_encode($duLieuTai));

        $chuKy = hash_hmac('sha256', "{$chuoiTieuDeBase64}.{$chuoiDuLieuBase64}", $this->khoaBiMat, true);
        $chuKyBase64 = $this->maHoaBase64Url($chuKy);

        return "{$chuoiTieuDeBase64}.{$chuoiDuLieuBase64}.{$chuKyBase64}";
    }

    /**
     * Sinh một chuỗi Refresh Token ngẫu nhiên có độ dài bảo mật cao (64 ký tự hex).
     */
    public function taoRefreshToken(): string
    {
        return bin2hex(random_bytes(32));
    }

    /**
     * Băm chuỗi Refresh Token bằng SHA-256 để lưu trữ an toàn trong cơ sở dữ liệu.
     */
    public function bamRefreshToken(string $chuoiRefreshToken): string
    {
        return hash('sha256', $chuoiRefreshToken);
    }

    /**
     * Giải mã và xác thực tính hợp lệ của Access Token.
     * Trả về dữ liệu payload nếu hợp lệ, ngược lại trả về null.
     */
    public function xacThucAccessToken(string $token): ?object
    {
        $cacPhan = explode('.', $token);
        if (count($cacPhan) !== 3) {
            return null;
        }

        [$chuoiTieuDeBase64, $chuoiDuLieuBase64, $chuKyBase64] = $cacPhan;

        // Đối soát chữ ký HMAC-SHA256
        $chuKyKyVong = hash_hmac('sha256', "{$chuoiTieuDeBase64}.{$chuoiDuLieuBase64}", $this->khoaBiMat, true);
        $chuKyThucTe = $this->giaiMaBase64Url($chuKyBase64);

        if (!hash_equals($chuKyKyVong, $chuKyThucTe)) {
            return null; // Chữ ký không khớp, token bị giả mạo
        }

        $duLieuJson = $this->giaiMaBase64Url($chuoiDuLieuBase64);
        $duLieuTai = json_decode($duLieuJson);

        if (!$duLieuTai || !isset($duLieuTai->exp) || !isset($duLieuTai->sub)) {
            return null;
        }

        // Kiểm tra thời hạn hiệu lực của token
        if (time() >= $duLieuTai->exp) {
            return null; // Token đã hết hạn
        }

        return $duLieuTai;
    }

    /**
     * Mã hóa chuỗi nhị phân theo chuẩn Base64URL (RFC 7515).
     */
    private function maHoaBase64Url(string $duLieu): string
    {
        return rtrim(strtr(base64_encode($duLieu), '+/', '-_'), '=');
    }

    /**
     * Giải mã chuỗi Base64URL về chuỗi nhị phân ban đầu.
     */
    private function giaiMaBase64Url(string $duLieu): string
    {
        $duLieu = strtr($duLieu, '-_', '+/');
        $soKyTuThieu = strlen($duLieu) % 4;
        if ($soKyTuThieu) {
            $duLieu .= str_repeat('=', 4 - $soKyTuThieu);
        }
        return base64_decode($duLieu);
    }
}
