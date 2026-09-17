<?php

namespace App\Http\Controllers;

use App\Models\DuAn;
use App\Models\NguoiDung;
use App\Services\DichVuAuditLog;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Validator;
use Illuminate\Support\Str;

/**
 * Controller quản lý các Dự án theo dõi CI/CD (Projects) và cấu hình Webhook.
 */
class DuAnController extends Controller
{
    protected DichVuAuditLog $dichVuAudit;

    public function __construct(DichVuAuditLog $dichVuAudit)
    {
        $this->dichVuAudit = $dichVuAudit;
    }

    /**
     * Lấy danh sách dự án có phân trang, tìm kiếm và thống kê trạng thái mới nhất.
     */
    public function danhSach(Request $yeuCau): JsonResponse
    {
        $tuKhoa = $yeuCau->query('search');
        $soLuongMoiTrang = (int) $yeuCau->query('per_page', 20);

        $truyVan = DuAn::with([
            'nhomSoHuu:id,name',
            'luotChayMoiNhat',
        ])->withCount(['danhSachCanhBaoDangMo as active_alerts_count']);

        if ($tuKhoa) {
            $truyVan->where(function ($sub) use ($tuKhoa) {
                $sub->where('name', 'ilike', "%{$tuKhoa}%")
                    ->orWhere('slug', 'ilike', "%{$tuKhoa}%");
            });
        }

        $ketQuaPhanTrang = $truyVan->orderBy('id', 'desc')->paginate($soLuongMoiTrang);

        return response()->json([
            'success' => true,
            'data' => $ketQuaPhanTrang->items(),
            'meta' => [
                'page' => $ketQuaPhanTrang->currentPage(),
                'per_page' => $ketQuaPhanTrang->perPage(),
                'total' => $ketQuaPhanTrang->total(),
                'total_pages' => $ketQuaPhanTrang->lastPage(),
            ],
        ]);
    }

    /**
     * Tạo một dự án mới kèm khóa bí mật Webhook tự động sinh (Yêu cầu vai trò Team Lead trở lên).
     */
    public function taoMoi(Request $yeuCau): JsonResponse
    {
        $kiemTraDuLieu = Validator::make($yeuCau->all(), [
            'name' => 'required|string|max:100',
            'description' => 'nullable|string',
            'repository_url' => 'nullable|url|max:255',
            'ci_provider' => 'nullable|string|in:github,gitlab,generic',
            'team_id' => 'nullable|exists:teams,id',
            'failure_threshold' => 'nullable|integer|min:1|max:20',
        ], [
            'name.required' => 'Tên dự án là bắt buộc.',
            'repository_url.url' => 'Đường dẫn repository phải là URL hợp lệ.',
        ]);

        if ($kiemTraDuLieu->fails()) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'VALIDATION_ERROR',
                    'message' => 'Dữ liệu dự án không hợp lệ.',
                    'details' => $kiemTraDuLieu->errors(),
                ],
            ], 422);
        }

        /** @var NguoiDung $nguoiDung */
        $nguoiDung = $yeuCau->attributes->get('nguoi_dung_hien_tai');

        $tenDuAn = $yeuCau->input('name');
        $chuoiSlug = Str::slug($tenDuAn);

        // Đảm bảo slug là duy nhất
        $soLuongTrung = DuAn::where('slug', 'like', "{$chuoiSlug}%")->count();
        if ($soLuongTrung > 0) {
            $chuoiSlug .= '-' . ($soLuongTrung + 1);
        }

        // Sinh khóa bí mật Webhook ngẫu nhiên
        $khoaBiMatWebhook = 'sec_' . bin2hex(random_bytes(20));

        $duAnMoi = DuAn::create([
            'team_id' => $yeuCau->input('team_id'),
            'name' => $tenDuAn,
            'slug' => $chuoiSlug,
            'description' => $yeuCau->input('description'),
            'repository_url' => $yeuCau->input('repository_url'),
            'ci_provider' => $yeuCau->input('ci_provider', 'github'),
            'webhook_secret' => $khoaBiMatWebhook,
            'failure_threshold' => (int) $yeuCau->input('failure_threshold', 3),
            'created_by' => $nguoiDung->id,
        ]);

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'tao_du_an',
            loaiThucThe: 'project',
            idThucThe: $duAnMoi->id,
            chiTiet: ['name' => $duAnMoi->name, 'slug' => $duAnMoi->slug]
        );

        return response()->json([
            'success' => true,
            'data' => array_merge($duAnMoi->toArray(), [
                'webhook_url' => "/api/v1/webhooks/projects/{$duAnMoi->id}",
            ]),
            'message' => 'Tạo dự án thành công.',
        ], 201);
    }

    /**
     * Xem thông tin chi tiết của một dự án.
     */
    public function chiTiet(int $id): JsonResponse
    {
        $duAn = DuAn::with([
            'nhomSoHuu:id,name',
            'nguoiTao:id,name,email',
            'luotChayMoiNhat',
        ])->withCount(['danhSachCanhBaoDangMo as active_alerts_count'])
          ->find($id);

        if (!$duAn) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy dự án yêu cầu.'],
            ], 404);
        }

        return response()->json([
            'success' => true,
            'data' => array_merge($duAn->toArray(), [
                'webhook_url' => "/api/v1/webhooks/projects/{$duAn->id}",
            ]),
        ]);
    }

    /**
     * Cập nhật thông tin cấu hình dự án.
     */
    public function capNhat(Request $yeuCau, int $id): JsonResponse
    {
        $duAn = DuAn::find($id);
        if (!$duAn) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy dự án.'],
            ], 404);
        }

        $duAn->update($yeuCau->only([
            'name',
            'description',
            'repository_url',
            'ci_provider',
            'failure_threshold',
            'team_id',
        ]));

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'cap_nhat_du_an',
            loaiThucThe: 'project',
            idThucThe: $duAn->id,
            chiTiet: $duAn->toArray()
        );

        return response()->json([
            'success' => true,
            'data' => $duAn,
            'message' => 'Cập nhật dự án thành công.',
        ]);
    }

    /**
     * Xóa dự án (Yêu cầu vai trò Admin).
     */
    public function xoa(int $id): JsonResponse
    {
        $duAn = DuAn::find($id);
        if (!$duAn) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy dự án.'],
            ], 404);
        }

        $tenDuAn = $duAn->name;
        $duAn->delete();

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'xoa_du_an',
            loaiThucThe: 'project',
            idThucThe: $id,
            chiTiet: ['name' => $tenDuAn]
        );

        return response()->json([
            'success' => true,
            'message' => 'Xóa dự án thành công.',
        ]);
    }

    /**
     * Sinh lại khóa bí mật Webhook (Webhooks Secret Regeneration).
     */
    public function sinhLaiSecret(int $id): JsonResponse
    {
        $duAn = DuAn::find($id);
        if (!$duAn) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy dự án.'],
            ], 404);
        }

        $khoaBiMatMoi = 'sec_' . bin2hex(random_bytes(20));
        $duAn->update(['webhook_secret' => $khoaBiMatMoi]);

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'sinh_lai_webhook_secret',
            loaiThucThe: 'project',
            idThucThe: $duAn->id,
            chiTiet: ['project_id' => $duAn->id]
        );

        return response()->json([
            'success' => true,
            'data' => [
                'webhook_secret' => $khoaBiMatMoi,
            ],
            'message' => 'Khóa bí mật Webhook đã được làm mới thành công.',
        ]);
    }
}
