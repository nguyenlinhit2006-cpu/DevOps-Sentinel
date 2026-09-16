<?php

namespace App\Http\Middleware;

use App\Services\DichVuAuditLog;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Middleware tự động ghi nhật ký kiểm toán (Audit Log) cho các yêu cầu thay đổi dữ liệu (POST, PUT, PATCH, DELETE).
 */
class GhiNhatKyHoatDong
{
    protected DichVuAuditLog $dichVuAudit;

    public function __construct(DichVuAuditLog $dichVuAudit)
    {
        $this->dichVuAudit = $dichVuAudit;
    }

    /**
     * Xử lý bắt các request thay đổi trạng thái và ghi log sau khi xử lý thành công.
     */
    public function handle(Request $yeuCau, Closure $tiepTuc): Response
    {
        $phanHoi = $tiepTuc($yeuCau);

        // Chỉ ghi nhật ký cho các thao tác thay đổi dữ liệu và xử lý thành công (mã 2xx)
        if (in_array($yeuCau->method(), ['POST', 'PUT', 'PATCH', 'DELETE']) && $phanHoi->isSuccessful()) {
            $duongDan = $yeuCau->path();
            $phuongThuc = $yeuCau->method();

            $chiTiet = [
                'method' => $phuongThuc,
                'path' => $duongDan,
                'status_code' => $phanHoi->getStatusCode(),
                'payload_keys' => array_keys($yeuCau->except(['password', 'password_confirmation', 'webhook_secret'])),
            ];

            $this->dichVuAudit->ghiNhatKy(
                hanhDong: strtolower("{$phuongThuc}_{$duongDan}"),
                loaiThucThe: 'http_request',
                idThucThe: null,
                chiTiet: $chiTiet,
                nguoiThaoTac: null,
                yeuCau: $yeuCau
            );
        }

        return $phanHoi;
    }
}
