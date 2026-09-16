<?php

namespace App\Http\Controllers;

use App\Models\NguoiDung;
use App\Services\DichVuJwt;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Validator;

/**
 * Controller phụ trách toàn bộ luồng xác thực và phân quyền (Authentication).
 * Bao gồm: Đăng ký tài khoản, đăng nhập, làm mới token (silent refresh), lấy thông tin hồ sơ và đăng xuất.
 */
class XacThucController extends Controller
{
    protected DichVuJwt $dichVuJwt;

    public function __construct(DichVuJwt $dichVuJwt)
    {
        $this->dichVuJwt = $dichVuJwt;
    }

    /**
     * Đăng ký tài khoản người dùng mới trong hệ thống.
     * Mặc định tài khoản mới nhận vai trò 'viewer'.
     */
    public function dangKy(Request $yeuCau): JsonResponse
    {
        $kiemTraDuLieu = Validator::make($yeuCau->all(), [
            'name' => 'required|string|max:100',
            'email' => 'required|string|email|max:255|unique:users,email',
            'password' => 'required|string|min:6',
            'password_confirmation' => 'required_with:password|same:password',
        ], [
            'name.required' => 'Họ tên là bắt buộc.',
            'email.required' => 'Địa chỉ email là bắt buộc.',
            'email.email' => 'Địa chỉ email không đúng định dạng.',
            'email.unique' => 'Địa chỉ email này đã được sử dụng.',
            'password.required' => 'Mật khẩu là bắt buộc.',
            'password.min' => 'Mật khẩu phải chứa ít nhất 6 ký tự.',
            'password_confirmation.same' => 'Xác nhận mật khẩu không trùng khớp.',
        ]);

        if ($kiemTraDuLieu->fails()) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'VALIDATION_ERROR',
                    'message' => 'Dữ liệu đăng ký không hợp lệ.',
                    'details' => $kiemTraDuLieu->errors(),
                ],
            ], 422);
        }

        $matKhauBam = password_hash($yeuCau->input('password'), PASSWORD_BCRYPT);

        $nguoiDungMoi = NguoiDung::create([
            'name' => $yeuCau->input('name'),
            'email' => strtolower(trim($yeuCau->input('email'))),
            'password_hash' => $matKhauBam,
            'role' => 'viewer',
        ]);

        $chuoiAccessToken = $this->dichVuJwt->taoAccessToken($nguoiDungMoi);
        $chuoiRefreshToken = $this->dichVuJwt->taoRefreshToken();

        // Lưu bản băm SHA-256 của Refresh Token vào cơ sở dữ liệu
        $nguoiDungMoi->refresh_token_hash = $this->dichVuJwt->bamRefreshToken($chuoiRefreshToken);
        $nguoiDungMoi->save();

        return response()->json([
            'success' => true,
            'data' => [
                'user' => [
                    'id' => $nguoiDungMoi->id,
                    'name' => $nguoiDungMoi->name,
                    'email' => $nguoiDungMoi->email,
                    'role' => $nguoiDungMoi->role,
                ],
                'access_token' => $chuoiAccessToken,
                'refresh_token' => $chuoiRefreshToken,
                'token_type' => 'Bearer',
                'expires_in' => 900,
            ],
            'message' => 'Đăng ký tài khoản thành công.',
        ], 201);
    }

    /**
     * Đăng nhập tài khoản bằng email và mật khẩu.
     * Cấp phát cặp Access Token và Refresh Token.
     */
    public function dangNhap(Request $yeuCau): JsonResponse
    {
        $kiemTraDuLieu = Validator::make($yeuCau->all(), [
            'email' => 'required|string|email',
            'password' => 'required|string',
        ], [
            'email.required' => 'Vui lòng nhập địa chỉ email.',
            'email.email' => 'Địa chỉ email không hợp lệ.',
            'password.required' => 'Vui lòng nhập mật khẩu.',
        ]);

        if ($kiemTraDuLieu->fails()) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'VALIDATION_ERROR',
                    'message' => 'Dữ liệu đăng nhập không hợp lệ.',
                    'details' => $kiemTraDuLieu->errors(),
                ],
            ], 422);
        }

        $diaChiEmail = strtolower(trim($yeuCau->input('email')));
        $matKhauNhap = $yeuCau->input('password');

        $nguoiDung = NguoiDung::where('email', $diaChiEmail)->first();

        if (!$nguoiDung || !password_verify($matKhauNhap, $nguoiDung->password_hash)) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'INVALID_CREDENTIALS',
                    'message' => 'Email hoặc mật khẩu không chính xác.',
                ],
            ], 401);
        }

        $chuoiAccessToken = $this->dichVuJwt->taoAccessToken($nguoiDung);
        $chuoiRefreshToken = $this->dichVuJwt->taoRefreshToken();

        // Cập nhật Refresh Token mới vào DB
        $nguoiDung->refresh_token_hash = $this->dichVuJwt->bamRefreshToken($chuoiRefreshToken);
        $nguoiDung->save();

        return response()->json([
            'success' => true,
            'data' => [
                'user' => [
                    'id' => $nguoiDung->id,
                    'name' => $nguoiDung->name,
                    'email' => $nguoiDung->email,
                    'role' => $nguoiDung->role,
                ],
                'access_token' => $chuoiAccessToken,
                'refresh_token' => $chuoiRefreshToken,
                'token_type' => 'Bearer',
                'expires_in' => 900,
            ],
            'message' => 'Đăng nhập thành công.',
        ]);
    }

    /**
     * Làm mới Access Token thông qua Refresh Token (Silent Refresh).
     */
    public function lamMoiToken(Request $yeuCau): JsonResponse
    {
        $kiemTraDuLieu = Validator::make($yeuCau->all(), [
            'refresh_token' => 'required|string',
        ], [
            'refresh_token.required' => 'Refresh token là bắt buộc để làm mới phiên làm việc.',
        ]);

        if ($kiemTraDuLieu->fails()) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'VALIDATION_ERROR',
                    'message' => 'Thiếu thông tin refresh token.',
                    'details' => $kiemTraDuLieu->errors(),
                ],
            ], 422);
        }

        $chuoiRefreshTokenNhan = $yeuCau->input('refresh_token');
        $chuoiBamKyVong = $this->dichVuJwt->bamRefreshToken($chuoiRefreshTokenNhan);

        $nguoiDung = NguoiDung::where('refresh_token_hash', $chuoiBamKyVong)->first();

        if (!$nguoiDung) {
            return response()->json([
                'success' => false,
                'error' => [
                    'code' => 'UNAUTHORIZED',
                    'message' => 'Refresh Token không hợp lệ hoặc phiên đăng nhập đã bị thu hồi.',
                ],
            ], 401);
        }

        // Cấp phát cặp token mới để xoay vòng token an toàn (Token Rotation)
        $chuoiAccessTokenMoi = $this->dichVuJwt->taoAccessToken($nguoiDung);
        $chuoiRefreshTokenMoi = $this->dichVuJwt->taoRefreshToken();

        $nguoiDung->refresh_token_hash = $this->dichVuJwt->bamRefreshToken($chuoiRefreshTokenMoi);
        $nguoiDung->save();

        return response()->json([
            'success' => true,
            'data' => [
                'access_token' => $chuoiAccessTokenMoi,
                'refresh_token' => $chuoiRefreshTokenMoi,
                'token_type' => 'Bearer',
                'expires_in' => 900,
            ],
            'message' => 'Làm mới Access Token thành công.',
        ]);
    }

    /**
     * Lấy thông tin tài khoản của người dùng đang đăng nhập hiện tại.
     */
    public function thongTinCaNhan(Request $yeuCau): JsonResponse
    {
        /** @var NguoiDung $nguoiDung */
        $nguoiDung = $yeuCau->attributes->get('nguoi_dung_hien_tai');

        return response()->json([
            'success' => true,
            'data' => [
                'user' => [
                    'id' => $nguoiDung->id,
                    'name' => $nguoiDung->name,
                    'email' => $nguoiDung->email,
                    'role' => $nguoiDung->role,
                    'created_at' => $nguoiDung->created_at?->toIso8601String(),
                ],
            ],
            'message' => 'Lấy thông tin người dùng thành công.',
        ]);
    }

    /**
     * Đăng xuất người dùng và thu hồi Refresh Token trên máy chủ.
     */
    public function dangXuat(Request $yeuCau): JsonResponse
    {
        /** @var NguoiDung $nguoiDung */
        $nguoiDung = $yeuCau->attributes->get('nguoi_dung_hien_tai');

        $nguoiDung->refresh_token_hash = null;
        $nguoiDung->save();

        return response()->json([
            'success' => true,
            'message' => 'Đăng xuất tài khoản thành công.',
        ]);
    }
}
