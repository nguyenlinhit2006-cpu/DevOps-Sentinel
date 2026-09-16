<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasOne;

/**
 * Model đại diện cho một lượt chạy pipeline CI/CD (pipeline_runs).
 */
class LuotChayPipeline extends Model
{
    protected $table = 'pipeline_runs';

    public $timestamps = false; // Bảng chỉ sử dụng trường created_at

    protected $fillable = [
        'project_id',
        'pipeline_name',
        'commit_hash',
        'commit_message',
        'branch',
        'author',
        'status',
        'trigger_event',
        'duration_seconds',
        'short_log',
        'external_url',
        'started_at',
        'finished_at',
        'created_at',
    ];

    protected function casts(): array
    {
        return [
            'started_at' => 'datetime',
            'finished_at' => 'datetime',
            'created_at' => 'datetime',
            'duration_seconds' => 'integer',
        ];
    }

    /**
     * Dự án chứa lượt chạy pipeline này.
     */
    public function duAn(): BelongsTo
    {
        return $this->belongsTo(DuAn::class, 'project_id');
    }

    /**
     * Cảnh báo kích hoạt bởi lần chạy này nếu có.
     */
    public function canhBao(): HasOne
    {
        return $this->hasOne(CanhBao::class, 'last_pipeline_run_id');
    }
}
