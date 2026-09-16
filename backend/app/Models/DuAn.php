<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\HasOne;

/**
 * Model đại diện cho một dự án cần theo dõi CI/CD (projects).
 */
class DuAn extends Model
{
    protected $table = 'projects';

    protected $fillable = [
        'team_id',
        'name',
        'slug',
        'description',
        'repository_url',
        'ci_provider',
        'webhook_secret',
        'failure_threshold',
        'created_by',
    ];

    /**
     * Nhóm sở hữu dự án này.
     */
    public function nhomSoHuu(): BelongsTo
    {
        return $this->belongsTo(Nhom::class, 'team_id');
    }

    /**
     * Người tạo dự án.
     */
    public function nguoiTao(): BelongsTo
    {
        return $this->belongsTo(NguoiDung::class, 'created_by');
    }

    /**
     * Toàn bộ lịch sử các lần chạy pipeline của dự án.
     */
    public function danhSachLuotChay(): HasMany
    {
        return $this->hasMany(LuotChayPipeline::class, 'project_id')->orderBy('created_at', 'desc');
    }

    /**
     * Lượt chạy pipeline gần nhất của dự án.
     */
    public function luotChayMoiNhat(): HasOne
    {
        return $this->hasOne(LuotChayPipeline::class, 'project_id')->latestOfMany('created_at');
    }

    /**
     * Danh sách tất cả các cảnh báo của dự án.
     */
    public function danhSachCanhBao(): HasMany
    {
        return $this->hasMany(CanhBao::class, 'project_id');
    }

    /**
     * Danh sách các cảnh báo đang ở trạng thái kích hoạt (active).
     */
    public function danhSachCanhBaoDangMo(): HasMany
    {
        return $this->hasMany(CanhBao::class, 'project_id')->where('status', 'active');
    }
}
