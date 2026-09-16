<?php

namespace App\Services;

use App\Models\CanhBao;
use App\Models\DuAn;
use App\Models\LuotChayPipeline;

/**
 * Lớp dịch vụ xử lý logic cảnh báo thông minh (Sentinel Alerting Rule).
 * Tự động phát hiện khi pipeline thất bại liên tiếp N lần để bật chuông cảnh báo.
 */
class DichVuCanhBao
{
    protected DichVuAuditLog $dichVuAudit;

    public function __construct(DichVuAuditLog $dichVuAudit)
    {
        $this->dichVuAudit = $dichVuAudit;
    }

    /**
     * Phân tích lượt chạy pipeline mới và kích hoạt hoặc đóng cảnh báo nếu cần.
     */
    public function kiemTraVaKichHoatCanhBao(DuAn $duAn, LuotChayPipeline $luotChayMoi): ?CanhBao
    {
        // 1. Trường hợp lượt chạy thất bại (failed)
        if ($luotChayMoi->status === 'failed') {
            $soLanLoiLienTiep = $this->tinhSoLanLoiLienTiep($duAn->id, $luotChayMoi->branch);
            $nguongThatBai = $duAn->failure_threshold;

            if ($soLanLoiLienTiep >= $nguongThatBai) {
                // Kiểm tra xem đã có cảnh báo nào đang mở cho dự án này chưa
                $canhBaoHienTai = CanhBao::where('project_id', $duAn->id)
                    ->where('status', 'active')
                    ->first();

                $tieuDeCanhBao = "Pipeline {$luotChayMoi->pipeline_name} thất bại {$soLanLoiLienTiep} lần liên tiếp trên nhánh {$luotChayMoi->branch}";
                $noiDungCanhBao = "Dự án {$duAn->name} đã ghi nhận {$soLanLoiLienTiep} lượt build thất bại liên tiếp. Trích đoạn lỗi gần nhất: " . ($luotChayMoi->short_log ?? 'Không có log chi tiết.');

                if ($canhBaoHienTai) {
                    // Cập nhật số lần lỗi và gắn với lượt chạy mới nhất
                    $canhBaoHienTai->update([
                        'last_pipeline_run_id' => $luotChayMoi->id,
                        'consecutive_failures' => $soLanLoiLienTiep,
                        'title' => $tieuDeCanhBao,
                        'message' => $noiDungCanhBao,
                    ]);
                    return $canhBaoHienTai;
                }

                // Tạo cảnh báo Sentinel mới
                $canhBaoMoi = CanhBao::create([
                    'project_id' => $duAn->id,
                    'last_pipeline_run_id' => $luotChayMoi->id,
                    'title' => $tieuDeCanhBao,
                    'message' => $noiDungCanhBao,
                    'consecutive_failures' => $soLanLoiLienTiep,
                    'status' => 'active',
                ]);

                // Ghi nhật ký kiểm toán hệ thống
                $this->dichVuAudit->ghiNhatKy(
                    hanhDong: 'kich_hoat_canh_bao',
                    loaiThucThe: 'alert',
                    idThucThe: $canhBaoMoi->id,
                    chiTiet: [
                        'project_id' => $duAn->id,
                        'project_name' => $duAn->name,
                        'branch' => $luotChayMoi->branch,
                        'consecutive_failures' => $soLanLoiLienTiep,
                    ]
                );

                return $canhBaoMoi;
            }
        }

        // 2. Trường hợp lượt chạy thành công (success)
        if ($luotChayMoi->status === 'success') {
            // Tự động đóng các cảnh báo đang mở của nhánh tương ứng khi hệ thống đã xanh trở lại
            $danhSachCanhBaoDangMo = CanhBao::where('project_id', $duAn->id)
                ->where('status', 'active')
                ->get();

            foreach ($danhSachCanhBaoDangMo as $canhBao) {
                $canhBao->update([
                    'status' => 'resolved',
                    'resolved_at' => now(),
                    'message' => $canhBao->message . ' (Hệ thống đã tự động phục hồi sau lượt build thành công).',
                ]);

                $this->dichVuAudit->ghiNhatKy(
                    hanhDong: 'tu_dong_phuc_hoi_canh_bao',
                    loaiThucThe: 'alert',
                    idThucThe: $canhBao->id,
                    chiTiet: [
                        'run_id' => $luotChayMoi->id,
                        'commit_hash' => $luotChayMoi->commit_hash,
                    ]
                );
            }
        }

        return null;
    }

    /**
     * Đếm số lần thất bại liên tiếp gần nhất của dự án trên nhánh chỉ định.
     */
    public function tinhSoLanLoiLienTiep(int $idDuAn, string $tenNhanh): int
    {
        $danhSachLuotChayGanNhat = LuotChayPipeline::where('project_id', $idDuAn)
            ->where('branch', $tenNhanh)
            ->orderBy('id', 'desc')
            ->limit(20)
            ->get();

        $soLanLoi = 0;
        foreach ($danhSachLuotChayGanNhat as $luotChay) {
            if ($luotChay->status === 'failed') {
                $soLanLoi++;
            } else {
                break; // Gặp lượt chạy không phải failed thì dừng đếm chuỗi liên tiếp
            }
        }

        return $soLanLoi;
    }
}
