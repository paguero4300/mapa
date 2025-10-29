@echo off
echo 🚀 Iniciando servidor en modo producción...
echo 📁 Ejecutando desde directorio raíz
echo.

REM Verificar si existe .env.local y renombrarlo temporalmente
if exist .env.local (
    echo 📦 Renombrando .env.local temporalmente para usar .env
    ren .env.local .env.local.backup
)

REM Cambiar al directorio backend y ejecutar
cd backend
go run main.go

REM Restaurar .env.local si existía
cd ..
if exist .env.local.backup (
    echo 📦 Restaurando .env.local
    ren .env.local.backup .env.local
)