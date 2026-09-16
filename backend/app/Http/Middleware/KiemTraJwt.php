<?php

namespace App\Http\Middleware;

use App\Models\NguoiDung;
use App\Services\DichVuJwt;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Middleware xác thực JSON Web Token từ tiêu đề Authorization của request.
 * Mục đích: Bảo vệ các API nghiệp vụ, định danh người dùng đăng nhập.
 */
class KiemTraJwt
{
    protected DichVuJwt $dichVuJwt;

    public function __construct(DichVuJwt $dichVuJwt)
    {
        $this->dichVuJwt = $dichVuJwt;
    }

    /**
     * Xử lý request kiểm tra token trước khi chuyển tiếp vào controller.
     */
    public function handle(Request $yeuCau, Closure $tiepTuc): Response
    {
        $tieuDeXacThuc = $yeuCau->header('Authorization');

        if (!$tieuDeXacThuc || !str_starts_with($tieuDeXacThuc, 'Bearer ')) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'UNAUTHORIZED',
                    'message' => 'Yêu cầu không có Access Token hoặc định dạng Authorization không đúng (Bearer <token>).',
                ],
            ], 401);
        }

        $chuoiToken = substr($tieuDeXacThuc, 7);
        $duLieuGiaiMa = $this->dichVuJwt->xacThucAccessToken($chuoiToken);

        if (!$duLieuGiaiMa) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'UNAUTHORIZED',
                    'message' => 'Access Token không hợp lệ hoặc đã hết hạn.',
                ],
            ], 401);
        }

        $nguoiDung = NguoiDung::find($duLieuGiaiMa->sub);

        if (!$nguoiDung) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'UNAUTHORIZED',
                    'message' => 'Tài khoản người dùng liên kết với token này không tồn tại trong hệ thống.',
                ],
            ], 401);
        }

        // Đính kèm đối tượng người dùng vào request để các controller sử dụng tiện lợi
        $yeuCau->attributes->set('nguoi_dung_hien_tai', $nguoiDung);

        return $tiepTuc($yeuCau);
    }
}
