@echo off
chcp 65001 >nul
title DevOps Sentinel - 1-Click Launcher (Docker Windows)

cls
echo ===============================================================================
echo              DEVOPS SENTINEL - BO KHOI CHAY 1-CHAM CHO WINDOWS
echo ===============================================================================
echo.

:: 1. Kiem tra Docker Desktop
echo [1/4] Dang kiem tra ket noi den Docker Desktop...
docker info >nul 2>&1
if errorlevel 1 (
    echo.
    echo [X] LOI: Khong the ket noi toi Docker Desktop!
    echo.
    echo Vui long thuc hien cac buoc sau:
    echo   1. Mo phan mem "Docker Desktop" tren may tinh Windows.
    echo   2. Cho den khi Docker Desktop hien thi trang thai chay (bieu tuong xanh).
    echo   3. Nhap dup chuot lai vao file run-windows.bat nay.
    echo.
    echo ===============================================================================
    pause
    exit /b 1
)
echo [OK] Docker Desktop dang hoat dong tot!
echo.

:: 2. Kiem tra Docker Image
echo [2/4] Dang kiem tra Docker Image 'devops-sentinel:latest'...
docker image inspect devops-sentinel:latest >nul 2>&1
if errorlevel 1 (
    echo Image 'devops-sentinel:latest' chua ton tai tren may.
    echo Dang tien hanh build tu Dockerfile (qua trinh nay chi mat 2-3 phut lan dau)...
    echo.
    docker build -t devops-sentinel:latest .
    if errorlevel 1 (
        echo.
        echo [X] LOI: Qua trinh build Docker Image that bai!
        echo Vui long kiem tra ket noi Internet va thu lai.
        echo.
        pause
        exit /b 1
    )
    echo [OK] Da build thanh cong Docker Image 'devops-sentinel:latest'!
) else (
    echo [OK] Image 'devops-sentinel:latest' da san sang!
)
echo.

:: 3. Kiem tra va Khoi chay Container
echo [3/4] Dang kiem tra Container 'devops-sentinel'...
docker inspect devops-sentinel >nul 2>&1
if errorlevel 1 (
    echo Dang tao va khoi chay Container moi...
    docker run -d -p 3000:3000 -p 8000:8000 -v devops_sentinel_data:/var/lib/postgresql/data --name devops-sentinel devops-sentinel:latest
    if errorlevel 1 (
        echo.
        echo [X] LOI: Khong the khoi chay Container! Co the cong 3000 hoac 8000 dang bi chiem dung.
        pause
        exit /b 1
    )
    echo [OK] Container da duoc tao va khoi chay thanh cong!
) else (
    set IS_RUNNING=false
    for /f "tokens=*" %%i in ('docker inspect -f "{{.State.Running}}" devops-sentinel 2^>nul') do set IS_RUNNING=%%i
    if "%IS_RUNNING%"=="true" (
        echo [OK] Container 'devops-sentinel' da dang chay san sang!
    ) else (
        echo Container dang dung. Dang khoi dong lai...
        docker start devops-sentinel
        if errorlevel 1 (
            echo Dang tao lai container moi...
            docker rm -f devops-sentinel >nul 2>&1
            docker run -d -p 3000:3000 -p 8000:8000 -v devops_sentinel_data:/var/lib/postgresql/data --name devops-sentinel devops-sentinel:latest
        )
        echo [OK] Container da duoc khoi dong lai!
    )
)
echo.

:: 4. Mo trinh duyet mac dinh
echo [4/4] Cho he thong san sang va mo trinh duyet...
timeout /t 3 /nobreak >nul 2>&1
start http://localhost:3000/#/login

echo.
echo ===============================================================================
echo           CHUC MUNG! HE THONG DEVOPS SENTINEL DA KHOI CHAY THANH CONG!
echo ===============================================================================
echo.
echo   * Giao dien WebAssembly (Frontend) : http://localhost:3000
echo   * May chu API Backend (Laravel)    : http://localhost:8000/api/health
echo.
echo   -----------------------------------------------------------------------------
echo   TAI KHOAN DANG NHAP THU NGHIEM (DA DUOC NAP SAN):
echo   1. Admin     : admin@devops-sentinel.local   ^| Mat khau: Sentinel@123456
echo   2. Team Lead : lead@devops-sentinel.local    ^| Mat khau: Sentinel@123456
echo   3. Viewer    : viewer@devops-sentinel.local  ^| Mat khau: Sentinel@123456
echo   -----------------------------------------------------------------------------
echo.
echo   CAC LENH QUAN LY HUU ICH (Mo CMD/PowerShell de dung):
echo   - Xem log truc tiep : docker logs -f devops-sentinel
echo   - Tam dung he thong : docker stop devops-sentinel   (hoac chay stop-windows.bat)
echo   - Chay lai he thong : docker start devops-sentinel  (hoac nhap dup run-windows.bat)
echo.
echo ===============================================================================
echo Ban co the dong cua so nay bat ky luc nao. Container van se tiep tuc chay ngam!
echo.
pause
