#!/bin/bash

echo "🚀 Iniciando servidor en modo producción..."
echo "📁 Ejecutando desde directorio raíz"
echo ""

# Verificar si existe .env.local y renombrarlo temporalmente
if [ -f .env.local ]; then
    echo "📦 Renombrando .env.local temporalmente para usar .env"
    mv .env.local .env.local.backup
fi

# Cambiar al directorio backend y ejecutar
cd backend
go run main.go

# Restaurar .env.local si existía
cd ..
if [ -f .env.local.backup ]; then
    echo "📦 Restaurando .env.local"
    mv .env.local.backup .env.local
fi