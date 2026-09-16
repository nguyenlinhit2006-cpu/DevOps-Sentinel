<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\BelongsToMany;

/**
 * Model đại diện cho nhóm làm việc (teams).
 * Cho phép phân chia các dự án và quản lý quyền hạn tập trung theo nhóm.
 */
class Nhom extends Model
{
    protected $table = 'teams';

    protected $fillable = [
        'name',
        'description',
        'created_by',
    ];

    /**
     * Người tạo ra nhóm làm việc này.
     */
    public function nguoiTao(): BelongsTo
    {
        return $this->belongsTo(NguoiDung::class, 'created_by');
    }

    /**
     * Danh sách các thành viên thuộc nhóm.
     */
    public function danhSachThanhVien(): BelongsToMany
    {
        return $this->belongsToMany(NguoiDung::class, 'team_members', 'team_id', 'user_id')
            ->withPivot('role');
    }

    /**
     * Danh sách các dự án thuộc quyền quản lý của nhóm.
     */
    public function danhSachDuAn(): HasMany
    {
        return $this->hasMany(DuAn::class, 'team_id');
    }
}
