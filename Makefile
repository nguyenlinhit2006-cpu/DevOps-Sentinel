.PHONY: env-up env-down migrate seed test dev-backend build-wasm dev-frontend

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

# Khởi chạy máy chủ Backend phục vụ phát triển (Cổng 8000)
dev-backend:
	cd backend && php artisan serve --port=8000

# Chạy toàn bộ kiểm thử backend
test:
	cd backend && php artisan test

# Biên dịch ứng dụng Go WebAssembly
build-wasm:
	cd frontend && GOOS=js GOARCH=wasm go build -o dist/main.wasm ./cmd/wasm

# Khởi chạy máy chủ phục vụ Frontend WebAssembly (Cổng 3000)
dev-frontend: build-wasm
	cd frontend && go run ./cmd/server
