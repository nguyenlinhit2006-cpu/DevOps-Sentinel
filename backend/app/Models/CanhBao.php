<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

/**
 * Model đại diện cho một cảnh báo Sentinel (alerts).
 */
class CanhBao extends Model
{
    protected $table = 'alerts';

    protected $fillable = [
        'project_id',
        'last_pipeline_run_id',
        'title',
        'message',
        'consecutive_failures',
        'status',
        'resolved_by',
        'resolved_at',
    ];

    protected function casts(): array
    {
        return [
            'consecutive_failures' => 'integer',
            'resolved_at' => 'datetime',
            'created_at' => 'datetime',
            'updated_at' => 'datetime',
        ];
    }

    /**
     * Dự án gặp cảnh báo.
     */
    public function duAn(): BelongsTo
    {
        return $this->belongsTo(DuAn::class, 'project_id');
    }

    /**
     * Lượt chạy pipeline gây ra cảnh báo.
     */
    public function luotChayLoi(): BelongsTo
    {
        return $this->belongsTo(LuotChayPipeline::class, 'last_pipeline_run_id');
    }

    /**
     * Người dùng đã xử lý / đóng cảnh báo này.
     */
    public function nguoiXuLy(): BelongsTo
    {
        return $this->belongsTo(NguoiDung::class, 'resolved_by');
    }
}
