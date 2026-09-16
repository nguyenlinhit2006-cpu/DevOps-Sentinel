<?php

namespace App\Http\Controllers;

use App\Models\NguoiDung;
use App\Models\Nhom;
use App\Services\DichVuAuditLog;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Validator;

/**
 * Controller quản lý Nhóm làm việc (Teams) và phân bổ thành viên nhóm.
 */
class NhomController extends Controller
{
    protected DichVuAuditLog $dichVuAudit;

    public function __construct(DichVuAuditLog $dichVuAudit)
    {
        $this->dichVuAudit = $dichVuAudit;
    }

    /**
     * Lấy danh sách tất cả các nhóm làm việc.
     */
    public function danhSach(): JsonResponse
    {
        $danhSachNhom = Nhom::withCount(['danhSachThanhVien as members_count', 'danhSachDuAn as projects_count'])
            ->with('nguoiTao:id,name,email')
            ->orderBy('id', 'asc')
            ->get();

        return response()->json([
            'success' => true,
            'data' => $danhSachNhom,
        ]);
    }

    /**
     * Tạo một nhóm làm việc mới (Yêu cầu vai trò Admin).
     */
    public function taoMoi(Request $yeuCau): JsonResponse
    {
        $kiemTraDuLieu = Validator::make($yeuCau->all(), [
            'name' => 'required|string|max:100',
            'description' => 'nullable|string',
        ], [
            'name.required' => 'Tên nhóm là bắt buộc.',
        ]);

        if ($kiemTraDuLieu->fails()) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'VALIDATION_ERROR',
                    'message' => 'Dữ liệu nhóm không hợp lệ.',
                    'details' => $kiemTraDuLieu->errors(),
                ],
            ], 422);
        }

        /** @var NguoiDung $nguoiDung */
        $nguoiDung = $yeuCau->attributes->get('nguoi_dung_hien_tai');

        $nhomMoi = Nhom::create([
            'name' => $yeuCau->input('name'),
            'description' => $yeuCau->input('description'),
            'created_by' => $nguoiDung->id,
        ]);

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'tao_nhom',
            loaiThucThe: 'team',
            idThucThe: $nhomMoi->id,
            chiTiet: ['name' => $nhomMoi->name]
        );

        return response()->json([
            'success' => true,
            'data' => $nhomMoi,
            'message' => 'Tạo nhóm làm việc thành công.',
        ], 201);
    }

    /**
     * Xem thông tin chi tiết của một nhóm bao gồm danh sách thành viên và dự án.
     */
    public function chiTiet(int $id): JsonResponse
    {
        $nhom = Nhom::with([
            'nguoiTao:id,name,email',
            'danhSachThanhVien:id,name,email,role',
            'danhSachDuAn:id,team_id,name,slug,ci_provider',
        ])->find($id);

        if (!$nhom) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy nhóm yêu cầu.'],
            ], 404);
        }

        return response()->json([
            'success' => true,
            'data' => $nhom,
        ]);
    }

    /**
     * Cập nhật thông tin nhóm.
     */
    public function capNhat(Request $yeuCau, int $id): JsonResponse
    {
        $nhom = Nhom::find($id);
        if (!$nhom) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy nhóm yêu cầu.'],
            ], 404);
        }

        $nhom->update($yeuCau->only(['name', 'description']));

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'cap_nhat_nhom',
            loaiThucThe: 'team',
            idThucThe: $nhom->id,
            chiTiet: $nhom->toArray()
        );

        return response()->json([
            'success' => true,
            'data' => $nhom,
            'message' => 'Cập nhật nhóm thành công.',
        ]);
    }

    /**
     * Xóa nhóm làm việc (Yêu cầu vai trò Admin).
     */
    public function xoa(int $id): JsonResponse
    {
        $nhom = Nhom::find($id);
        if (!$nhom) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy nhóm yêu cầu.'],
            ], 404);
        }

        $tenNhom = $nhom->name;
        $nhom->delete();

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'xoa_nhom',
            loaiThucThe: 'team',
            idThucThe: $id,
            chiTiet: ['name' => $tenNhom]
        );

        return response()->json([
            'success' => true,
            'message' => 'Xóa nhóm thành công.',
        ]);
    }

    /**
     * Thêm hoặc cập nhật thành viên vào nhóm.
     */
    public function themThanhVien(Request $yeuCau, int $id): JsonResponse
    {
        $nhom = Nhom::find($id);
        if (!$nhom) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy nhóm.'],
            ], 404);
        }

        $kiemTraDuLieu = Validator::make($yeuCau->all(), [
            'user_id' => 'required|exists:users,id',
            'role' => 'required|in:team_lead,viewer',
        ]);

        if ($kiemTraDuLieu->fails()) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'VALIDATION_ERROR', 'message' => 'Dữ liệu thành viên không hợp lệ.', 'details' => $kiemTraDuLieu->errors()],
            ], 422);
        }

        $userId = $yeuCau->input('user_id');
        $vaiTro = $yeuCau->input('role');

        $nhom->danhSachThanhVien()->syncWithoutDetaching([$userId => ['role' => $vaiTro]]);

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'them_thanh_vien_nhom',
            loaiThucThe: 'team',
            idThucThe: $id,
            chiTiet: ['user_id' => $userId, 'role' => $vaiTro]
        );

        return response()->json([
            'success' => true,
            'message' => 'Đã thêm thành viên vào nhóm thành công.',
        ]);
    }

    /**
     * Xóa thành viên khỏi nhóm.
     */
    public function xoaThanhVien(int $id, int $userId): JsonResponse
    {
        $nhom = Nhom::find($id);
        if (!$nhom) {
            return response()->json([
                'success' => false,
                'error' => ['code' => 'NOT_FOUND', 'message' => 'Không tìm thấy nhóm.'],
            ], 404);
        }

        $nhom->danhSachThanhVien()->detach($userId);

        $this->dichVuAudit->ghiNhatKy(
            hanhDong: 'xoa_thanh_vien_nhom',
            loaiThucThe: 'team',
            idThucThe: $id,
            chiTiet: ['user_id' => $userId]
        );

        return response()->json([
            'success' => true,
            'message' => 'Đã xóa thành viên khỏi nhóm thành công.',
        ]);
    }
}
