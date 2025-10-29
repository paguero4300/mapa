package middleware

import (
	"log"
	"net/http"
	"strings"
	"traccar-login/config"
	"traccar-login/services"

	"github.com/gin-gonic/gin"
)

// isAPIRequest verifica si la petición es a un endpoint de API
func isAPIRequest(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/api/")
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("🔒 [MIDDLEWARE] ===== INICIANDO VERIFICACIÓN DE AUTENTICACIÓN =====")
		log.Printf("🔒 [MIDDLEWARE] URL: %s", c.Request.URL.Path)
		log.Printf("🔒 [MIDDLEWARE] Method: %s", c.Request.Method)
		log.Printf("🔒 [MIDDLEWARE] User-Agent: %s", c.GetHeader("User-Agent"))
		log.Printf("🔒 [MIDDLEWARE] Referer: %s", c.GetHeader("Referer"))

		// PRIMERO: Verificar si hay una sesión global activa
		globalSession := services.GetGlobalSession()
		if globalSession != nil {
			log.Printf("🌐 [MIDDLEWARE] Sesión global encontrada, verificando...")

			// Verificar sesión con el servicio global
			user, err := globalSession.GetSession()
			if err == nil && user != nil && user.ID != 0 {
				log.Printf("✅ [MIDDLEWARE] Sesión global válida para usuario: %s (ID: %d)", user.Name, user.ID)

				// Guardar información del usuario y el servicio en el contexto
				c.Set("userId", user.ID)
				c.Set("userEmail", user.Email)
				c.Set("userName", user.Name)
				c.Set("userAdmin", user.Administrator)
				c.Set("user", user)
				c.Set("traccarService", globalSession)

				log.Printf("✅ [MIDDLEWARE] ===== AUTENTICACIÓN GLOBAL EXITOSA - CONTINUANDO =====")
				c.Next()
				return
			} else {
				log.Printf("❌ [MIDDLEWARE] Sesión global inválida: %v", err)
				// Limpiar sesión global inválida
				services.SetGlobalSession(nil)
			}
		}

		// SEGUNDO: Si no hay sesión global, intentar con cookies del cliente
		log.Printf("🍪 [MIDDLEWARE] No hay sesión global, verificando cookies del cliente...")
		cookies := c.Request.Cookies()
		log.Printf("🍪 [MIDDLEWARE] Cookies recibidas: %d", len(cookies))

		for i, cookie := range cookies {
			log.Printf("🍪 [MIDDLEWARE] Cookie %d: %s=%s (Path: %s, Domain: %s, HttpOnly: %v, Secure: %v)",
				i+1, cookie.Name, cookie.Value, cookie.Path, cookie.Domain, cookie.HttpOnly, cookie.Secure)
		}

		// Verificar que hay cookies de sesión
		if len(cookies) == 0 {
			log.Printf("❌ [MIDDLEWARE] ===== NO HAY COOKIES NI SESIÓN GLOBAL =====")
			// Si es una petición API, devolver JSON error
			if isAPIRequest(c) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}

		log.Printf("🔄 [MIDDLEWARE] Creando TraccarService con cookies...")
		// Crear TraccarService con las cookies del cliente
		traccarService := services.NewTraccarService(cfg.TraccarURL)
		traccarService = traccarService.WithCookies(cookies)
		log.Printf("🔄 [MIDDLEWARE] TraccarService creado con %d cookies", len(cookies))

		log.Printf("🔍 [MIDDLEWARE] Verificando sesión con Traccar...")
		// Verificar sesión con Traccar
		user, err := traccarService.GetSession()
		if err != nil {
			log.Printf("❌ [MIDDLEWARE] ===== ERROR VERIFICANDO SESIÓN =====")
			log.Printf("❌ [MIDDLEWARE] Error detallado: %v", err)
			// Si es una petición API, devolver JSON error
			if isAPIRequest(c) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesión inválida"})
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}

		log.Printf("🔍 [MIDDLEWARE] Respuesta de Traccar recibida, verificando usuario...")
		// Verificar que el usuario tiene datos válidos
		if user == nil {
			log.Printf("❌ [MIDDLEWARE] ===== USUARIO ES NIL =====")
			// Si es una petición API, devolver JSON error
			if isAPIRequest(c) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario inválido"})
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}

		if user.ID == 0 {
			log.Printf("❌ [MIDDLEWARE] ===== USUARIO ID ES 0 =====")
			// Si es una petición API, devolver JSON error
			if isAPIRequest(c) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario inválido"})
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}

		log.Printf("✅ [MIDDLEWARE] ===== USUARIO VÁLIDO ENCONTRADO =====")
		log.Printf("✅ [MIDDLEWARE] Usuario autenticado: ID=%d, Name=%s, Email=%s, Admin=%v",
			user.ID, user.Name, user.Email, user.Administrator)

		// Establecer esta sesión como global para futuras peticiones
		services.SetGlobalSession(traccarService)

		log.Printf("💾 [MIDDLEWARE] Guardando datos en contexto...")
		// Guardar información del usuario y el servicio en el contexto
		c.Set("userId", user.ID)
		c.Set("userEmail", user.Email)
		c.Set("userName", user.Name)
		c.Set("userAdmin", user.Administrator)
		c.Set("user", user)
		c.Set("traccarService", traccarService)

		log.Printf("✅ [MIDDLEWARE] ===== AUTENTICACIÓN EXITOSA - CONTINUANDO =====")
		c.Next()
		log.Printf("✅ [MIDDLEWARE] ===== MIDDLEWARE COMPLETADO =====")
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// Permitir orígenes específicos para desarrollo y producción
		allowedOrigins := []string{
			"http://localhost:8080",
			"http://127.0.0.1:8080",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}

		// Verificar si el origen está permitido
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		// Si no hay origen o está permitido, configurar CORS
		if origin == "" || isAllowed {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				// Para peticiones sin Origin (como herramientas CLI), permitir localhost
				c.Header("Access-Control-Allow-Origin", "http://localhost:8080")
			}
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
