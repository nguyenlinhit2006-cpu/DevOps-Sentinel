<?php

namespace App\Models;

use Illuminate\Foundation\Auth\User as Authenticatable;
use Illuminate\Notifications\Notifiable;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\BelongsToMany;

/**
 * Model đại diện cho người dùng hệ thống.
 * Quản lý thông tin tài khoản, vai trò phân quyền (admin, team_lead, viewer) và token xác thực.
 */
class NguoiDung extends Authenticatable
{
    use Notifiable;

    /**
     * Tên bảng trong cơ sở dữ liệu PostgreSQL.
     */
    protected $table = 'users';

    /**
     * Các trường dữ liệu được phép gán hàng loạt (mass assignable).
     */
    protected $fillable = [
        'name',
        'email',
        'password_hash',
        'role',
        'refresh_token_hash',
    ];

    /**
     * Các trường dữ liệu ẩn khi serialize sang JSON.
     */
    protected $hidden = [
        'password_hash',
        'refresh_token_hash',
    ];

    /**
     * Định dạng ép kiểu dữ liệu cho các trường.
     */
    protected function casts(): array
    {
        return [
            'created_at' => 'datetime',
            'updated_at' => 'datetime',
        ];
    }

    /**
     * Lấy mật khẩu phục vụ cơ chế xác thực của Laravel.
     */
    public function getAuthPassword(): string
    {
        return $this->password_hash;
    }

    /**
     * Thiết lập quan hệ danh sách các nhóm mà người dùng tham gia.
     */
    public function danhSachNhom(): BelongsToMany
    {
        return $this->belongsToMany(Nhom::class, 'team_members', 'user_id', 'team_id')
            ->withPivot('role')
            ->withTimestamps();
    }

    /**
     * Thiết lập quan hệ các dự án do người dùng trực tiếp tạo ra.
     */
    public function danhSachDuAnDaTao(): HasMany
    {
        return $this->hasMany(DuAn::class, 'created_by');
    }

    /**
     * Kiểm tra xem người dùng có phải quản trị viên tối cao (Admin) hay không.
     */
    public function laQuanTriVien(): bool
    {
        return $this->role === 'admin';
    }

    /**
     * Kiểm tra xem người dùng có quyền trưởng nhóm (Team Lead) trở lên hay không.
     */
    public function laTruongNhom(): bool
    {
        return in_array($this->role, ['admin', 'team_lead']);
    }
}
