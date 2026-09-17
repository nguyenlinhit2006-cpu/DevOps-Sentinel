# ==============================================================================
# DevOps Sentinel - Đóng gói Trọn gói Tất cả trong một (All-in-One Docker Image)
# Bao gồm: Frontend Go WebAssembly + Backend Laravel 11 API + PostgreSQL 16
# ==============================================================================

# ------------------------------------------------------------------------------
# Giai đoạn 1: Biên dịch Frontend Go WebAssembly & Máy chủ tĩnh
# ------------------------------------------------------------------------------
FROM golang:1.23-alpine AS wasm-builder

WORKDIR /build

COPY frontend/go.mod ./
COPY frontend/ .

# Biên dịch mã nguồn Go sang WebAssembly (main.wasm)
RUN GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o dist/main.wasm ./cmd/wasm

# Biên dịch máy chủ phục vụ file tĩnh Go
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

# ------------------------------------------------------------------------------
# Giai đoạn 2: Cài đặt Dependencies Backend PHP (Composer)
# ------------------------------------------------------------------------------
FROM composer:2 AS composer-builder

WORKDIR /build

COPY backend/composer.json backend/composer.lock ./

RUN composer install \
    --no-dev \
    --no-interaction \
    --prefer-dist \
    --optimize-autoloader \
    --no-scripts \
    --ignore-platform-reqs

# ------------------------------------------------------------------------------
# Giai đoạn 3: Runtime Image Chính (Ubuntu 24.04 LTS: PHP 8.3 & PostgreSQL 16)
# ------------------------------------------------------------------------------
FROM ubuntu:24.04 AS final

ENV DEBIAN_FRONTEND=noninteractive \
    TZ=Asia/Ho_Chi_Minh \
    PGDATA=/var/lib/postgresql/data

# Cài đặt PHP 8.3, PostgreSQL 16 và các thư viện cần thiết
RUN apt-get update && apt-get install -y --no-install-recommends \
    php-cli \
    php-pgsql \
    php-mbstring \
    php-xml \
    php-curl \
    php-zip \
    php-bcmath \
    postgresql \
    postgresql-client \
    curl \
    ca-certificates \
    bash \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Thiết lập thư mục ứng dụng
WORKDIR /app

# Sao chép mã nguồn Backend
COPY backend /app/backend
COPY --from=composer-builder /build/vendor /app/backend/vendor

# Sao chép mã nguồn Frontend và file binary đã biên dịch
COPY frontend /app/frontend
COPY --from=wasm-builder /build/dist/main.wasm /app/frontend/dist/main.wasm
COPY --from=wasm-builder /build/bin/server /app/frontend/bin/server

# Sao chép kịch bản khởi chạy entrypoint
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Cấu hình lưu trữ dữ liệu bền vững
VOLUME ["/var/lib/postgresql/data"]

# Khai báo các cổng dịch vụ
# 3000: Giao diện người dùng WebAssembly
# 8000: Máy chủ RESTful API Backend
# 5432: Cơ sở dữ liệu PostgreSQL
EXPOSE 3000 8000 5432

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
