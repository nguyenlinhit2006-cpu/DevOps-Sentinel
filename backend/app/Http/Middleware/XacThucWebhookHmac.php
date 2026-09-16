<?php

namespace App\Http\Middleware;

use App\Models\DuAn;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * Middleware xác thực chữ ký HMAC SHA-256 hoặc Secret Token của Webhook đẩy từ CI/CD Providers.
 * Hỗ trợ GitHub Actions (X-Hub-Signature-256), GitLab CI (X-Gitlab-Token), và Generic (X-Sentinel-Signature).
 */
class XacThucWebhookHmac
{
    /**
     * Kiểm tra tính hợp lệ của chữ ký trước khi chuyển tiếp sang WebhookController.
     */
    public function handle(Request $yeuCau, Closure $tiepTuc): Response
    {
        $idDuAn = $yeuCau->route('id') ?? $yeuCau->route('projectId');

        if (!$idDuAn) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'NOT_FOUND',
                    'message' => 'Không xác định được ID dự án trong đường dẫn webhook.',
                ],
            ], 404);
        }

        $duAn = DuAn::find($idDuAn);

        if (!$duAn) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'NOT_FOUND',
                    'message' => 'Dự án cấu hình nhận webhook không tồn tại.',
                ],
            ], 404);
        }

        $khoaBiMat = $duAn->webhook_secret;
        $noiDungTho = $yeuCau->getContent();

        // 1. Kiểm tra chữ ký GitHub Actions
        $chuKyGithub = $yeuCau->header('X-Hub-Signature-256');
        if ($chuKyGithub) {
            $chuKyKyVong = 'sha256=' . hash_hmac('sha256', $noiDungTho, $khoaBiMat);
            if (!hash_equals($chuKyKyVong, $chuKyGithub)) {
                return $this->traVeLoiChuKyKhongHopLe();
            }
            $yeuCau->attributes->set('du_an_webhook', $duAn);
            return $tiepTuc($yeuCau);
        }

        // 2. Kiểm tra mã bí mật GitLab CI Token
        $tokenGitlab = $yeuCau->header('X-Gitlab-Token');
        if ($tokenGitlab) {
            if (!hash_equals($khoaBiMat, $tokenGitlab)) {
                return $this->traVeLoiChuKyKhongHopLe();
            }
            $yeuCau->attributes->set('du_an_webhook', $duAn);
            return $tiepTuc($yeuCau);
        }

        // 3. Kiểm tra chữ ký Generic Sentinel Webhook
        $chuKyGeneric = $yeuCau->header('X-Sentinel-Signature');
        if ($chuKyGeneric) {
            $chuKyKyVong = 'sha256=' . hash_hmac('sha256', $noiDungTho, $khoaBiMat);
            if (!hash_equals($chuKyKyVong, $chuKyGeneric)) {
                return $this->traVeLoiChuKyKhongHopLe();
            }
            $yeuCau->attributes->set('du_an_webhook', $duAn);
            return $tiepTuc($yeuCau);
        }

        // Nếu không cung cấp bất kỳ tiêu đề xác thực nào
        return response()->json([
            'success' => false,
            'error' => [
                'code' => 'INVALID_SIGNATURE',
                'message' => 'Thiếu tiêu đề xác thực chữ ký Webhook (X-Hub-Signature-256, X-Gitlab-Token hoặc X-Sentinel-Signature).',
            ],
        ], 401);
    }

    /**
     * Trả về phản hồi lỗi chữ ký không hợp lệ theo chuẩn Error Envelope.
     */
    private function traVeLoiChuKyKhongHopLe(): Response
    {
        return response()->json([
            'success' => false,
            'error' => [
                'code' => 'INVALID_SIGNATURE',
                'message' => 'Chữ ký Webhook HMAC-SHA256 hoặc Secret Token không hợp lệ.',
            ],
        ], 401);
    }
}
