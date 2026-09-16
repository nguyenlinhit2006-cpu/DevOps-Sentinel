<?php

namespace App\Http\Controllers;

use App\Models\DuAn;
use App\Models\LuotChayPipeline;
use App\Services\DichVuAuditLog;
use App\Services\DichVuCanhBao;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Carbon;

/**
 * Controller tiếp nhận và chuẩn hóa sự kiện Webhook từ các CI/CD Providers (GitHub, GitLab, Generic).
 */
class WebhookController extends Controller
{
    protected DichVuCanhBao $dichVuCanhBao;
    protected DichVuAuditLog $dichVuAudit;

    public function __construct(DichVuCanhBao $dichVuCanhBao, DichVuAuditLog $dichVuAudit)
    {
        $this->dichVuCanhBao = $dichVuCanhBao;
        $this->dichVuAudit = $dichVuAudit;
    }

    /**
     * Tiếp nhận sự kiện webhook từ CI Provider sau khi đã qua middleware kiểm tra chữ ký HMAC.
     */
    public function tiepNhan(Request $yeuCau, int $id): JsonResponse
    {
        /** @var DuAn $duAn */
        $duAn = $yeuCau->attributes->get('du_an_webhook') ?? DuAn::findOrFail($id);

        $duLieuPayload = $yeuCau->all();
        $duLieuChuanHoa = $this->chuanHoaDuLieuWebhook($duLieuPayload, $duAn->ci_provider);

        // Giới hạn độ dài short_log tối đa 64KB (65535 ký tự)
        if (isset($duLieuChuanHoa['short_log']) && strlen($duLieuChuanHoa['short_log']) > 65000) {
            $duLieuChuanHoa['short_log'] = substr($duLieuChuanHoa['short_log'], 0, 65000) . "\n...[Log đã được cắt gọn]";
        }

        $luotChay = LuotChayPipeline::create(array_merge($duLieuChuanHoa, [
            'project_id' => $duAn->id,
            'created_at' => Carbon::now(),
        ]));

        // Kích hoạt dịch vụ Sentinel Alert phân tích lỗi liên tiếp
        $canhBaoKichHoat = $this->dichVuCanhBao->kiemTraVaKichHoatCanhBao($duAn, $luotChay);

        // Ghi nhật ký kiểm toán hệ thống
        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'tiep_nhan_webhook',
            loaiThucThe: 'pipeline_run',
            idThucThe: $luotChay->id,
            chiTiet: [
                'project_id' => $duAn->id,
                'status' => $luotChay->status,
                'branch' => $luotChay->branch,
                'commit_hash' => $luotChay->commit_hash,
                'alert_triggered' => ($canhBaoKichHoat !== null),
            ]
        );

        return response()->json([
            'success' => true,
            'data' => [
                'run_id' => $luotChay->id,
                'status' => 'processed',
                'alert_triggered' => ($canhBaoKichHoat !== null),
            ],
            'message' => 'Webhook đã được tiếp nhận và xử lý thành công.',
        ]);
    }

    /**
     * Chuẩn hóa payload từ các định dạng khác nhau (GitHub, GitLab, Generic) về cấu trúc thống nhất.
     */
    private function chuanHoaDuLieuWebhook(array $payload, string $nhaCungCap): array
    {
        // 1. Nếu payload gửi theo chuẩn GitHub Actions
        if (isset($payload['workflow_run'])) {
            $chayWorkflow = $payload['workflow_run'];
            $ketLuan = $chayWorkflow['conclusion'] ?? $chayWorkflow['status'] ?? 'queued';

            $trangThai = match ($ketLuan) {
                'success' => 'success',
                'failure', 'timed_out' => 'failed',
                'cancelled' => 'cancelled',
                'in_progress' => 'running',
                default => 'queued',
            };

            return [
                'pipeline_name' => $chayWorkflow['name'] ?? 'default',
                'commit_hash' => $chayWorkflow['head_sha'] ?? 'unknown',
                'commit_message' => $chayWorkflow['head_commit']['message'] ?? $chayWorkflow['display_title'] ?? null,
                'branch' => $chayWorkflow['head_branch'] ?? 'main',
                'author' => $chayWorkflow['actor']['login'] ?? null,
                'status' => $trangThai,
                'trigger_event' => $chayWorkflow['event'] ?? 'push',
                'duration_seconds' => isset($chayWorkflow['run_started_at'], $chayWorkflow['updated_at'])
                    ? Carbon::parse($chayWorkflow['run_started_at'])->diffInSeconds(Carbon::parse($chayWorkflow['updated_at']))
                    : 0,
                'short_log' => $payload['short_log'] ?? null,
                'external_url' => $chayWorkflow['html_url'] ?? null,
                'started_at' => isset($chayWorkflow['run_started_at']) ? Carbon::parse($chayWorkflow['run_started_at']) : null,
                'finished_at' => isset($chayWorkflow['updated_at']) ? Carbon::parse($chayWorkflow['updated_at']) : null,
            ];
        }

        // 2. Nếu payload gửi theo chuẩn GitLab CI Pipeline
        if (isset($payload['object_attributes'])) {
            $thuocTinh = $payload['object_attributes'];
            $trangThaiGitlab = $thuocTinh['status'] ?? 'pending';

            $trangThai = match ($trangThaiGitlab) {
                'success' => 'success',
                'failed' => 'failed',
                'canceled' => 'cancelled',
                'running' => 'running',
                default => 'queued',
            };

            return [
                'pipeline_name' => $payload['project']['name'] ?? 'gitlab-ci',
                'commit_hash' => $thuocTinh['sha'] ?? 'unknown',
                'commit_message' => $payload['commit']['message'] ?? null,
                'branch' => $thuocTinh['ref'] ?? 'main',
                'author' => $payload['user']['name'] ?? null,
                'status' => $trangThai,
                'trigger_event' => $thuocTinh['source'] ?? 'push',
                'duration_seconds' => (int) ($thuocTinh['duration'] ?? 0),
                'short_log' => $payload['short_log'] ?? null,
                'external_url' => $thuocTinh['url'] ?? null,
                'started_at' => isset($thuocTinh['created_at']) ? Carbon::parse($thuocTinh['created_at']) : null,
                'finished_at' => isset($thuocTinh['finished_at']) ? Carbon::parse($thuocTinh['finished_at']) : null,
            ];
        }

        // 3. Chuẩn Generic REST JSON Schema (Theo API Contract Phase 2)
        $trangThaiNhap = strtolower($payload['status'] ?? 'queued');
        $trangThaiHopLe = in_array($trangThaiNhap, ['queued', 'running', 'success', 'failed', 'cancelled'])
            ? $trangThaiNhap
            : 'queued';

        return [
            'pipeline_name' => $payload['pipeline_name'] ?? 'default',
            'commit_hash' => $payload['commit_hash'] ?? bin2hex(random_bytes(16)),
            'commit_message' => $payload['commit_message'] ?? null,
            'branch' => $payload['branch'] ?? 'main',
            'author' => $payload['author'] ?? 'system',
            'status' => $trangThaiHopLe,
            'trigger_event' => $payload['trigger_event'] ?? 'push',
            'duration_seconds' => (int) ($payload['duration_seconds'] ?? 0),
            'short_log' => $payload['short_log'] ?? null,
            'external_url' => $payload['external_url'] ?? null,
            'started_at' => isset($payload['started_at']) ? Carbon::parse($payload['started_at']) : null,
            'finished_at' => isset($payload['finished_at']) ? Carbon::parse($payload['finished_at']) : null,
        ];
    }
}
