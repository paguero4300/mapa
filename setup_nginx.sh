#!/bin/bash

echo "🌐 Configurando Nginx como reverse proxy para MAPA..."

# Paso 1: Crear archivo de configuración de Nginx
echo "📝 Creando configuración de Nginx..."
sudo tee /etc/nginx/sites-available/mapa > /dev/null << 'EOF'
server {
    listen 80;
    server_name 161.132.47.157;

    # Servir archivos estáticos directamente
    location /static/ {
        alias /var/www/html/mapa/frontend/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Proxy para la aplicación Go en puerto 8080
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
EOF

# Paso 2: Habilitar el sitio
echo "🔗 Habilitando el sitio mapa..."
sudo ln -sf /etc/nginx/sites-available/mapa /etc/nginx/sites-enabled/

# Paso 3: Eliminar el sitio default
echo "🗑️ Eliminando sitio default de Nginx..."
sudo rm -f /etc/nginx/sites-enabled/default

# Paso 4: Probar configuración de Nginx
echo "🧪 Probando configuración de Nginx..."
sudo nginx -t

if [ $? -eq 0 ]; then
    echo "✅ Configuración de Nginx válida"
    
    # Paso 5: Reiniciar Nginx
    echo "🔄 Reiniciando Nginx..."
    sudo systemctl restart nginx
    
    # Paso 6: Verificar estado
    echo "📊 Verificando estado de Nginx..."
    sudo systemctl status nginx --no-pager
    
    echo ""
    echo "🎉 Configuración completada!"
    echo "🌐 Ahora puedes acceder a tu aplicación en: http://161.132.47.157/"
    echo ""
    echo "🔍 Si aún no funciona, verifica:"
    echo "   - Que el servicio mapa esté corriendo: sudo systemctl status mapa"
    echo "   - Logs de Nginx: sudo journalctl -u nginx -f"
    echo "   - Logs de la aplicación: sudo journalctl -u mapa -f"
else
    echo "❌ Error en la configuración de Nginx"
    echo "🔍 Revisa la configuración con: sudo nginx -t"
fi