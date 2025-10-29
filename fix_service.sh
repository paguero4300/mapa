#!/bin/bash

echo "🔧 Iniciando proceso de reparación del servicio mapa..."

# Paso 1: Detener el servicio
echo "🛑 Deteniendo el servicio mapa..."
sudo systemctl stop mapa

# Paso 2: Navegar al backend y recompilar
echo "📦 Recompilando la aplicación..."
cd /var/www/html/mapa/backend
go build -o ../mapa-app .

# Paso 3: Configurar permisos
echo "🔐 Configurando permisos..."
cd /var/www/html/mapa
sudo chown -R www-data:www-data /var/www/html/mapa
sudo chmod +x /var/www/html/mapa/mapa-app

# Paso 4: Iniciar el servicio
echo "🚀 Iniciando el servicio mapa..."
sudo systemctl start mapa

# Paso 5: Verificar el estado
echo "📊 Verificando el estado del servicio..."
sudo systemctl status mapa

# Paso 6: Verificar que la aplicación responde
echo "🌐 Verificando que la aplicación responde..."
sleep 3
curl -s http://localhost:8080 || echo "❌ La aplicación no responde en localhost:8080"

echo ""
echo "✅ Proceso completado. Si el servicio sigue fallando, revisa los logs con:"
echo "   sudo journalctl -u mapa -f"