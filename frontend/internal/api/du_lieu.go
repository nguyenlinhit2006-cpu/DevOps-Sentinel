//go:build js && wasm

package api

/**
 * Cấu trúc dữ liệu số liệu tóm tắt cho Bảng điều khiển (Dashboard).
 */
type ThongKeBangDieuKhien struct {
	TotalProjects      int                 `json:"total_projects"`
	TotalRuns          int                 `json:"total_runs"`
	TotalRunsToday     int                 `json:"total_runs_today"`
	SuccessRunsCount   int                 `json:"success_runs_count"`
	FailedRunsCount    int                 `json:"failed_runs_count"`
	SuccessRatePercent float64             `json:"success_rate_percent"`
	ActiveAlertsCount  int                 `json:"active_alerts_count"`
	RecentRuns         []ThongTinLuotChay  `json:"recent_runs"`
}

/**
 * Thông tin chi tiết một lượt chạy pipeline CI/CD.
 */
type ThongTinLuotChay struct {
	ID              int64             `json:"id"`
	ProjectID       int64             `json:"project_id"`
	PipelineName    string            `json:"pipeline_name"`
	CommitHash      string            `json:"commit_hash"`
	CommitMessage   string            `json:"commit_message"`
	Branch          string            `json:"branch"`
	Author          string            `json:"author"`
	Status          string            `json:"status"`
	TriggerEvent    string            `json:"trigger_event"`
	DurationSeconds int               `json:"duration_seconds"`
	ShortLog        string            `json:"short_log"`
	ExternalURL     string            `json:"external_url"`
	StartedAt       string            `json:"started_at"`
	FinishedAt      string            `json:"finished_at"`
	CreatedAt       string            `json:"created_at"`
	DuAn            *ThongTinDuAnNho  `json:"du_an"`
}

/**
 * Cấu trúc tóm tắt dự án lồng trong thông tin lượt chạy.
 */
type ThongTinDuAnNho struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

/**
 * Cấu trúc nhóm chủ quản thu nhỏ.
 */
type ThongTinNhomNho struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

/**
 * Cấu trúc thông tin một dự án theo dõi CI/CD.
 */
type ThongTinDuAn struct {
	ID               int64             `json:"id"`
	TeamID           *int64            `json:"team_id"`
	Name             string            `json:"name"`
	Slug             string            `json:"slug"`
	Description      string            `json:"description"`
	RepositoryURL    string            `json:"repository_url"`
	CIProvider       string            `json:"ci_provider"`
	WebhookSecret    string            `json:"webhook_secret"`
	FailureThreshold int               `json:"failure_threshold"`
	NhomSoHuu        *ThongTinNhomNho  `json:"nhom_so_huu"`
	LuotChayMoiNhat  *ThongTinLuotChay `json:"luot_chay_moi_nhat"`
	ActiveAlertsCount int              `json:"active_alerts_count"`
	CreatedAt        string            `json:"created_at"`
}

/**
 * Cấu trúc phân trang trả về từ Backend.
 */
type PhanTrangMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

/**
 * Cấu trúc thông tin người tạo dự án.
 */
type ThongTinNguoiTao struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

/**
 * Cấu trúc chi tiết đầy đủ của một dự án kèm thông tin cấu hình webhook.
 */
type ThongTinDuAnChiTiet struct {
	ID                int64             `json:"id"`
	TeamID            *int64            `json:"team_id"`
	Name              string            `json:"name"`
	Slug              string            `json:"slug"`
	Description       string            `json:"description"`
	RepositoryURL     string            `json:"repository_url"`
	CIProvider        string            `json:"ci_provider"`
	WebhookSecret     string            `json:"webhook_secret"`
	WebhookURL        string            `json:"webhook_url"`
	FailureThreshold  int               `json:"failure_threshold"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
	NhomSoHuu         *ThongTinNhomNho  `json:"nhom_so_huu"`
	NguoiTao          *ThongTinNguoiTao `json:"nguoi_tao"`
	LuotChayMoiNhat   *ThongTinLuotChay `json:"luot_chay_moi_nhat"`
	ActiveAlertsCount int               `json:"active_alerts_count"`
}

