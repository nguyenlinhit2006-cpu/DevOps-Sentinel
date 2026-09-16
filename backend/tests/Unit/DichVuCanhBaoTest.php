<?php

namespace Tests\Unit;

use App\Models\DuAn;
use App\Models\LuotChayPipeline;
use App\Services\DichVuAuditLog;
use App\Services\DichVuCanhBao;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

/**
 * Kiểm thử đơn vị cho thuật toán đếm chuỗi thất bại liên tiếp của DichVuCanhBao.
 */
class DichVuCanhBaoTest extends TestCase
{
    use RefreshDatabase;

    protected DichVuCanhBao $dichVuCanhBao;

    protected function setUp(): void
    {
        parent::setUp();
        $this->dichVuCanhBao = new DichVuCanhBao(new DichVuAuditLog());
    }

    /**
     * Kiểm thử hàm tinhSoLanLoiLienTiep đếm chính xác chuỗi fail gần nhất.
     */
    public function test_tinh_so_lan_loi_lien_tiep_chinh_xac(): void
    {
        $duAn = DuAn::create([
            'name' => 'Dự Án Unit Test Alert',
            'slug' => 'du-an-unit-test-alert',
            'webhook_secret' => 'secret_unit_test',
        ]);

        // Thêm các lượt chạy theo thứ tự thời gian tăng dần:
        // Run 1: failed (cũ nhất)
        // Run 2: success
        // Run 3: failed
        // Run 4: failed (mới nhất)
        LuotChayPipeline::create(['project_id' => $duAn->id, 'branch' => 'main', 'status' => 'failed', 'commit_hash' => 'h1']);
        LuotChayPipeline::create(['project_id' => $duAn->id, 'branch' => 'main', 'status' => 'success', 'commit_hash' => 'h2']);
        LuotChayPipeline::create(['project_id' => $duAn->id, 'branch' => 'main', 'status' => 'failed', 'commit_hash' => 'h3']);
        LuotChayPipeline::create(['project_id' => $duAn->id, 'branch' => 'main', 'status' => 'failed', 'commit_hash' => 'h4']);

        // Nhánh main: 2 lần fail gần nhất liên tiếp -> Kết quả phải bằng 2
        $soLanLoiMain = $this->dichVuCanhBao->tinhSoLanLoiLienTiep($duAn->id, 'main');
        $this->assertEquals(2, $soLanLoiMain);

        // Nhánh develop: Chưa có lượt chạy nào -> Kết quả phải bằng 0
        $soLanLoiDevelop = $this->dichVuCanhBao->tinhSoLanLoiLienTiep($duAn->id, 'develop');
        $this->assertEquals(0, $soLanLoiDevelop);
    }
}
