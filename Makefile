.PHONY: env-up env-down migrate seed test dev-backend

# Khởi chạy các container nền tảng (PostgreSQL 16 & Mailpit)
env-up:
	docker compose up -d

# Dừng các container
env-down:
	docker compose down

# Chạy migration cơ sở dữ liệu
migrate:
	cd backend && php artisan migrate

# Nạp dữ liệu mẫu
seed:
	cd backend && php artisan db:seed

# Khởi chạy máy chủ Backend phục vụ phát triển
dev-backend:
	cd backend && php artisan serve --port=8000

# Chạy toàn bộ kiểm thử
test:
	cd backend && php artisan test
