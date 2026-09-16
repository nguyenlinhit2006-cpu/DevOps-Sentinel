<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

/**
 * Model đại diện cho một bản ghi kiểm toán hệ thống (audit_logs).
 */
class NhatKyHoatDong extends Model
{
    protected $table = 'audit_logs';

    public $timestamps = false; // Chỉ dùng trường created_at

    protected $fillable = [
        'user_id',
        'action',
        'entity_type',
        'entity_id',
        'ip_address',
        'user_agent',
        'details',
        'created_at',
    ];

    protected function casts(): array
    {
        return [
            'details' => 'array',
            'created_at' => 'datetime',
        ];
    }

    /**
     * Người dùng thực hiện hành vi kiểm toán này.
     */
    public function nguoiThaoTac(): BelongsTo
    {
        return $this->belongsTo(NguoiDung::class, 'user_id');
    }
}
