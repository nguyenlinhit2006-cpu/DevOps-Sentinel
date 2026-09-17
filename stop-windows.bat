@echo off
chcp 65001 >nul
title DevOps Sentinel - Dung Container

cls
echo ===============================================================================
echo                DEVOPS SENTINEL - TAM DUNG CONTAINER
echo ===============================================================================
echo.
echo Dang tam dung container 'devops-sentinel'...
docker stop devops-sentinel
if errorlevel 1 (
    echo.
    echo Container chua duoc khoi chay hoac da dung truoc do.
) else (
    echo.
    echo [OK] Da tam dung container 'devops-sentinel' thanh cong!
)
echo.
echo ===============================================================================
echo De khoi dong lai, ban chi can nhap dup chuot vao file run-windows.bat.
echo ===============================================================================
echo.
pause
