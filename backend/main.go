package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"traccar-login/config"
	"traccar-login/handlers"
	"traccar-login/middleware"
	"traccar-login/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error cargando configuración:", err)
	}

	// Inicializar servicios
	traccarService := services.NewTraccarService(cfg.TraccarURL)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(traccarService)
	dashboardHandler := handlers.NewDashboardHandler()
	devicesHandler := handlers.NewDevicesHandler(traccarService)
	mapHandler := handlers.NewMapHandler(traccarService)
	geofencesHandler := handlers.NewGeofencesHandler(traccarService)

	// Configurar Gin
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORSMiddleware())
	r.LoadHTMLGlob("../frontend/templates/*")
	r.Static("/static", "../frontend/static")

	// Ruta principal con verificación de autenticación
	r.GET("/", func(c *gin.Context) {
		log.Printf("🏠 [MAIN] ===== ACCESO A RUTA PRINCIPAL ===== ")
		log.Printf("🏠 [MAIN] User-Agent: %s", c.GetHeader("User-Agent"))
		log.Printf("🏠 [MAIN] Referer: %s", c.GetHeader("Referer"))
		
		// Verificar si el usuario ya está autenticado
		cookies := c.Request.Cookies()
		log.Printf("🍪 [MAIN] Cookies recibidas: %d", len(cookies))
		for i, cookie := range cookies {
			log.Printf("🍪 [MAIN] Cookie %d: %s=%s", i+1, cookie.Name, cookie.Value)
		}
		
		if len(cookies) > 0 {
			// Crear servicio Traccar con las cookies del cliente
			traccarService := services.NewTraccarService(cfg.TraccarURL)
			traccarService = traccarService.WithCookies(cookies)
			
			// Verificar sesión
			user, err := traccarService.GetSession()
			if err == nil && user != nil {
				log.Printf("✅ [MAIN] Usuario ya autenticado: %s, redirigiendo al dashboard", user.Name)
				c.Redirect(http.StatusFound, "/dashboard")
				return
			}
			log.Printf("❌ [MAIN] Sesión no válida: %v", err)
		}
		
		log.Printf("📝 [MAIN] Mostrando página de login")
		c.HTML(http.StatusOK, "login.html", nil)
		log.Printf("📝 [MAIN] ===== LOGIN RENDERIZADO ===== ")
	})

	// Rutas API públicas
	api := r.Group("/api")
	{
		api.POST("/login", authHandler.Login)
		api.POST("/logout", authHandler.Logout)
		api.GET("/session", authHandler.GetSession) // Verificación de sesión sin middleware
		
		// Endpoint de debug
		api.GET("/debug", func(c *gin.Context) {
			log.Printf("🔍 [DEBUG] ===== ENDPOINT DE DEBUG =====")
			cookies := c.Request.Cookies()
			log.Printf("🔍 [DEBUG] Cookies: %d", len(cookies))
			for i, cookie := range cookies {
				log.Printf("🔍 [DEBUG] Cookie %d: %s=%s", i+1, cookie.Name, cookie.Value)
			}
			
			if len(cookies) > 0 {
				traccarService := services.NewTraccarService(cfg.TraccarURL)
				traccarService = traccarService.WithCookies(cookies)
				user, err := traccarService.GetSession()
				if err != nil {
					c.JSON(http.StatusOK, gin.H{
						"cookies": len(cookies),
						"error": err.Error(),
						"status": "session_invalid",
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"cookies": len(cookies),
					"user": user,
					"status": "authenticated",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"cookies": 0,
					"status": "no_cookies",
				})
			}
		})
		
		// Endpoint de debug para rutas
		api.GET("/routes", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Rutas disponibles",
				"routes": []string{
					"GET /api/geofences",
					"PUT /api/geofences/:id",
					"DELETE /api/geofences/:id",
					"PUT /api/geofences/:id/visibility",
				},
				"server_status": "running",
			})
		})
		
		// Endpoint de debug para visibilidad (sin autenticación)
		api.PUT("/test-visibility/:id", func(c *gin.Context) {
			log.Printf("🧪 [DEBUG] ===== TEST VISIBILITY ENDPOINT =====")
			log.Printf("🧪 [DEBUG] Method: %s", c.Request.Method)
			log.Printf("🧪 [DEBUG] URL: %s", c.Request.URL.String())
			log.Printf("🧪 [DEBUG] Headers: %v", c.Request.Header)
			
			// Leer body
			body, _ := io.ReadAll(c.Request.Body)
			log.Printf("🧪 [DEBUG] Request body: %s", string(body))
			
			// Intentar parsear JSON
			var testReq struct {
				Visible bool `json:"visible"`
			}
			
			if err := json.Unmarshal(body, &testReq); err != nil {
				log.Printf("❌ [DEBUG] Error parseando JSON: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido: " + err.Error()})
				return
			}
			
			log.Printf("✅ [DEBUG] JSON parseado correctamente: visible=%t", testReq.Visible)
			c.JSON(http.StatusOK, gin.H{
				"message": "Test exitoso",
				"received_visible": testReq.Visible,
				"geofence_id": c.Param("id"),
			})
		})
		
		// Página de test integrada
		api.GET("/test-page", func(c *gin.Context) {
			htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>Test Visibility API</title>
</head>
<body>
    <h1>Test Visibility API</h1>
    
    <h2>Test 1: Debug Endpoint (sin auth)</h2>
    <button onclick="testDebug()">Test Debug</button>
    <div id="debugResult"></div>
    
    <h2>Test 2: Real Endpoint (con auth)</h2>
    <button onclick="testReal()">Test Real</button>
    <div id="realResult"></div>

    <script>
    async function testDebug() {
        const resultDiv = document.getElementById('debugResult');
        try {
            console.log('Testing debug endpoint...');
            const response = await fetch('/api/test-visibility/40', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ visible: false })
            });
            
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText);
            }
            
            const result = await response.json();
            resultDiv.innerHTML = '<h3>✅ Debug Success!</h3><pre>' + JSON.stringify(result, null, 2) + '</pre>';
        } catch (error) {
            resultDiv.innerHTML = '<h3>❌ Debug Failed</h3><p>' + error.message + '</p>';
        }
    }
    
    async function testReal() {
        const resultDiv = document.getElementById('realResult');
        try {
            console.log('Testing real endpoint...');
            const response = await fetch('/api/geofences/40/visibility', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ visible: false })
            });
            
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText);
            }
            
            const result = await response.json();
            resultDiv.innerHTML = '<h3>✅ Real Success!</h3><pre>' + JSON.stringify(result, null, 2) + '</pre>';
        } catch (error) {
            resultDiv.innerHTML = '<h3>❌ Real Failed</h3><p>' + error.message + '</p>';
        }
    }
    </script>
</body>
</html>`
			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, htmlContent)
		})
	}

	// Rutas protegidas con middleware de autenticación
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.GET("/dashboard", dashboardHandler.Dashboard)
		protected.GET("/devices", devicesHandler.DevicesPage)
		protected.GET("/map", mapHandler.MapPage)
		protected.GET("/geofences", geofencesHandler.GeofencesPage)
	}
	
	// API protegida para dispositivos
	devicesAPI := r.Group("/api/devices")
	devicesAPI.Use(middleware.AuthMiddleware(cfg))
	{
		devicesAPI.GET("", devicesHandler.GetDevices)
		devicesAPI.POST("", devicesHandler.CreateDevice)
		devicesAPI.PUT("/:id", devicesHandler.UpdateDevice)
		devicesAPI.DELETE("/:id", devicesHandler.DeleteDevice)
	}
	
	// API protegida para mapa
	mapAPI := r.Group("/api/map")
	mapAPI.Use(middleware.AuthMiddleware(cfg))
	{
		mapAPI.GET("/data", mapHandler.GetMapData)
		mapAPI.GET("/positions", mapHandler.GetPositions)
		mapAPI.GET("/device/:id/position", mapHandler.GetDevicePosition)
		mapAPI.GET("/geofences", mapHandler.GetGeofences)
	}
	
	// API protegida para geocercas
	log.Printf("🔷 [MAIN] Registrando rutas de geocercas...")
	geofencesAPI := r.Group("/api/geofences")
	geofencesAPI.Use(middleware.AuthMiddleware(cfg))
	{
		geofencesAPI.GET("", geofencesHandler.GetGeofences)
		geofencesAPI.PUT("/:id", geofencesHandler.UpdateGeofence)
		geofencesAPI.DELETE("/:id", geofencesHandler.DeleteGeofence)
		geofencesAPI.PUT("/:id/visibility", geofencesHandler.UpdateGeofenceVisibility)
	}
	log.Printf("✅ [MAIN] Rutas de geocercas registradas:")
	log.Printf("✅ [MAIN]   GET    /api/geofences")
	log.Printf("✅ [MAIN]   PUT    /api/geofences/:id")
	log.Printf("✅ [MAIN]   DELETE /api/geofences/:id")
	log.Printf("✅ [MAIN]   PUT    /api/geofences/:id/visibility")

	// Iniciar servidor
	log.Printf("Servidor iniciado en puerto %s", cfg.ServerPort)
	log.Printf("URL de Traccar: %s", cfg.TraccarURL)
	log.Printf("🔧 Usando autenticación basada en cookies de Traccar")
	log.Fatal(r.Run(":" + cfg.ServerPort))
}
