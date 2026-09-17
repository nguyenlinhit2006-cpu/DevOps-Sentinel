#!/usr/bin/env bash
set -e

echo "🚀 [DevOps Sentinel] Khởi động Container All-in-One..."

PGDATA="${PGDATA:-/var/lib/postgresql/data}"
DB_USER="${DB_USERNAME:-sentinel_user}"
DB_PASS="${DB_PASSWORD:-sentinel_password}"
DB_NAME="${DB_DATABASE:-devops_sentinel}"

# 1. Khởi tạo cụm dữ liệu PostgreSQL nếu chưa có
mkdir -p "$PGDATA" /var/log/postgresql /var/run/postgresql
chown -R postgres:postgres "$PGDATA" /var/log/postgresql /var/run/postgresql
chmod 777 /var/run/postgresql

if [ ! -s "$PGDATA/PG_VERSION" ]; then
    echo "🐘 [PostgreSQL] Khởi tạo cụm dữ liệu mới tại $PGDATA..."
    su - postgres -c "/usr/lib/postgresql/16/bin/initdb -D '$PGDATA' --auth-local=trust --auth-host=scram-sha-256"

    # Cho phép kết nối cục bộ và từ xa
    echo "listen_addresses = '*'" >> "$PGDATA/postgresql.conf"
    echo "host all all 0.0.0.0/0 scram-sha-256" >> "$PGDATA/pg_hba.conf"
    echo "host all all 127.0.0.1/32 trust" >> "$PGDATA/pg_hba.conf"
    echo "host all all all trust" >> "$PGDATA/pg_hba.conf"
fi

# Khởi chạy dịch vụ PostgreSQL
echo "🐘 [PostgreSQL] Đang khởi động tiến trình cơ sở dữ liệu..."
su - postgres -c "/usr/lib/postgresql/16/bin/pg_ctl -D '$PGDATA' -l /var/log/postgresql/postgresql.log start"

# Đợi PostgreSQL sẵn sàng
until su - postgres -c "pg_isready -h 127.0.0.1 -p 5432" > /dev/null 2>&1; do
    echo "⏳ [PostgreSQL] Đang đợi PostgreSQL khởi động hoàn tất..."
    sleep 1
done
echo "✅ [PostgreSQL] Dịch vụ cơ sở dữ liệu đã sẵn sàng!"

# Khởi tạo Role và Database nếu chưa tồn tại
su - postgres -c "psql -tc \"SELECT 1 FROM pg_roles WHERE rolname = '$DB_USER'\" | grep -q 1 || psql -c \"CREATE ROLE $DB_USER WITH LOGIN PASSWORD '$DB_PASS' CREATEDB;\""
su - postgres -c "psql -tc \"SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'\" | grep -q 1 || psql -c \"CREATE DATABASE $DB_NAME OWNER $DB_USER;\""

# 2. Cấu hình Backend Laravel
cd /app/backend

if [ ! -f .env ]; then
    cp .env.example .env
fi

# Cập nhật kết nối DB trong .env
sed -i "s/^DB_CONNECTION=.*/DB_CONNECTION=pgsql/" .env
sed -i "s/^DB_HOST=.*/DB_HOST=127.0.0.1/" .env
sed -i "s/^DB_PORT=.*/DB_PORT=5432/" .env
sed -i "s/^DB_DATABASE=.*/DB_DATABASE=$DB_NAME/" .env
sed -i "s/^DB_USERNAME=.*/DB_USERNAME=$DB_USER/" .env
sed -i "s/^DB_PASSWORD=.*/DB_PASSWORD=$DB_PASS/" .env

# Làm sạch bootstrap cache cũ nếu có
rm -f bootstrap/cache/*.php

# Sinh APP_KEY nếu chưa có
if ! grep -q "^APP_KEY=base64:" .env; then
    php artisan key:generate --force
fi

# Tự động phát hiện các packages khả dụng trong production
php artisan package:discover --ansi 2>/dev/null || true

# Phân quyền thư mục lưu trữ
mkdir -p storage/framework/{sessions,views,cache} storage/logs bootstrap/cache
chmod -R 777 storage bootstrap/cache

# Chạy migration & nạp dữ liệu mẫu
echo "🔄 [Backend] Chạy migration tạo bảng cơ sở dữ liệu..."
php artisan migrate --force

echo "🌱 [Backend] Kiểm tra và nạp dữ liệu mẫu..."
php artisan db:seed --force 2>/dev/null || true

# 3. Khởi chạy Backend API Server (Cổng 8000)
echo "⚡ [Backend] Khởi chạy máy chủ API tại http://0.0.0.0:8000..."
php artisan serve --host=0.0.0.0 --port=8000 &
BACKEND_PID=$!

# 4. Khởi chạy Frontend Go WebAssembly Server (Cổng 3000)
cd /app/frontend
echo "🌐 [Frontend] Khởi chạy máy chủ WebAssembly tại http://0.0.0.0:3000..."
./bin/server &
FRONTEND_PID=$!

echo "=========================================================="
echo "🎉 [DevOps Sentinel] Toàn bộ hệ thống đã khởi chạy thành công!"
echo "   • Giao diện WebAssembly (Frontend): http://localhost:3000"
echo "   • Máy chủ RESTful API (Backend):   http://localhost:8000"
echo "   • Cơ sở dữ liệu PostgreSQL:        localhost:5432"
echo "   • Tài khoản quản trị Admin:        admin@devops-sentinel.local"
echo "   • Mật khẩu quản trị mặc định:      Sentinel@123456"
echo "=========================================================="

# Bắt tín hiệu dừng từ Docker
dung_he_thong() {
    echo "🛑 [DevOps Sentinel] Đang dừng các dịch vụ..."
    kill -TERM "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
    su - postgres -c "/usr/lib/postgresql/16/bin/pg_ctl -D '$PGDATA' stop" || true
    exit 0
}

trap dung_he_thong SIGTERM SIGINT

wait -n
