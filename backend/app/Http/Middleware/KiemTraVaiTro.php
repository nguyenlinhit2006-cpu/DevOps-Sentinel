<?php

namespace App\Http\Middleware;

use App\Models\NguoiDung;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Middleware kiểm tra và phân quyền người dùng theo vai trò (RBAC).
 * Mục đích: Đảm bảo người dùng có vai trò phù hợp (admin, team_lead, viewer) trước khi thực hiện hành động.
 */
class KiemTraVaiTro
{
    /**
     * Bảng ánh xạ cấp bậc quyền hạn (Càng cao càng nhiều quyền).
     */
    private array $capBacVaiTro = [
        'admin' => 3,
        'team_lead' => 2,
        'viewer' => 1,
    ];

    /**
     * Xử lý kiểm tra vai trò của người dùng trong request.
     *
     * @param Request $yeuCau
     * @param Closure $tiepTuc
     * @param string $vaiTroYeuCau Vai trò tối thiểu yêu cầu (admin, team_lead, hoặc viewer)
     */
    public function handle(Request $yeuCau, Closure $tiepTuc, string $vaiTroYeuCau): Response
    {
        /** @var NguoiDung|null $nguoiDung */
        $nguoiDung = $yeuCau->attributes->get('nguoi_dung_hien_tai');

        if (!$nguoiDung) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'UNAUTHORIZED',
                    'message' => 'Chưa xác thực danh tính người dùng.',
                ],
            ], 401);
        }

        $capDoNguoiDung = $this->capBacVaiTro[$nguoiDung->role] ?? 0;
        $capDoYeuCau = $this->capBacVaiTro[$vaiTroYeuCau] ?? 99;

        if ($capDoNguoiDung < $capDoYeuCau) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'FORBIDDEN',
                    'message' => "Bạn không có quyền thực hiện hành động này. Yêu cầu vai trò tối thiểu: {$vaiTroYeuCau}.",
                ],
            ], 403);
        }

        return $tiepTuc($yeuCau);
    }
}
