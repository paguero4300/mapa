package handlers

import (
	"log"
	"net/http"
	"time"
	"traccar-login/models"
	"traccar-login/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	traccarService *services.TraccarService
}

func NewAuthHandler(traccarService *services.TraccarService) *AuthHandler {
	return &AuthHandler{
		traccarService: traccarService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	log.Printf("🔐 [AUTH] Iniciando proceso de login")

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ [AUTH] Error en validación de datos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de login inválidos: " + err.Error()})
		return
	}

	log.Printf("📧 [AUTH] Intentando login para email: %s", req.Email)

	// Autenticar con Traccar
	loginResponse, err := h.traccarService.Login(req.Email, req.Password)
	if err != nil {
		log.Printf("❌ [AUTH] Error en login: %v", err)
		// Determinar el tipo de error para respuesta más específica
		if err.Error() == "credenciales inválidas" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email o contraseña incorrectos"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando con el servidor de autenticación"})
		}
		return
	}

	log.Printf("✅ [AUTH] Login exitoso, estableciendo %d cookies", len(loginResponse.Cookies))

	// Establecer cookies en la respuesta del cliente
	log.Printf("🍪 [AUTH] ===== ESTABLECIENDO COOKIES EN EL CLIENTE =====")
	for i, cookie := range loginResponse.Cookies {
		log.Printf("🍪 [AUTH] Procesando cookie %d: %s=%s", i+1, cookie.Name, cookie.Value)
		log.Printf("🍪 [AUTH] Cookie original - Domain: %s, Path: %s, HttpOnly: %v, Secure: %v",
			cookie.Domain, cookie.Path, cookie.HttpOnly, cookie.Secure)

		// Configuración de cookie CORREGIDA para desarrollo local
		clientCookie := &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     "/",                  // Path raíz para toda la aplicación
			Domain:   "localhost",          // Especificar localhost explícitamente
			MaxAge:   86400,                // 24 horas
			Secure:   false,                // false para HTTP en desarrollo
			HttpOnly: true,                 // true para seguridad (JavaScript no necesita acceso directo)
			SameSite: http.SameSiteLaxMode, // Lax para navegación normal
		}

		http.SetCookie(c.Writer, clientCookie)

		log.Printf("🍪 [AUTH] Cookie FORZADA establecida: %s=%s", clientCookie.Name, clientCookie.Value)
		log.Printf("🍪 [AUTH] Configuración FORZADA - Path: %s, Domain: %s, MaxAge: %d, HttpOnly: %v, Secure: %v, SameSite: %v",
			clientCookie.Path, clientCookie.Domain, clientCookie.MaxAge, clientCookie.HttpOnly, clientCookie.Secure, clientCookie.SameSite)
	}
	log.Printf("🍪 [AUTH] ===== COOKIES ESTABLECIDAS COMPLETAMENTE =====")

	// Verificar que el usuario tiene los datos necesarios
	if loginResponse.User == nil {
		log.Printf("❌ [AUTH] Usuario nulo en respuesta de login")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error procesando datos de usuario"})
		return
	}

	log.Printf("✅ [AUTH] Login completado exitosamente para usuario: %s (ID: %d)",
		loginResponse.User.Name, loginResponse.User.ID)

	// Responder con usuario (sin token, usamos cookies)
	c.JSON(http.StatusOK, models.LoginResponse{
		User:  loginResponse.User,
		Token: "", // No usamos JWT, usamos cookies de Traccar
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	log.Printf("🚪 [AUTH] Iniciando proceso de logout")

	// Crear un servicio Traccar con las cookies del cliente para el logout
	traccarService := h.traccarService

	// Obtener cookies del cliente
	clientCookies := c.Request.Cookies()
	log.Printf("🍪 [AUTH] Cookies del cliente para logout: %d", len(clientCookies))

	if len(clientCookies) > 0 {
		// Crear servicio con las cookies del cliente
		traccarService = h.traccarService.WithCookies(clientCookies)
		log.Printf("🔄 [AUTH] Servicio Traccar creado con cookies del cliente")
	}

	// Cerrar sesión en Traccar
	err := traccarService.Logout()
	if err != nil {
		log.Printf("⚠️ [AUTH] Error cerrando sesión en Traccar: %v", err)
		// No fallar por esto, continuamos limpiando cookies locales
	} else {
		log.Printf("✅ [AUTH] Sesión cerrada exitosamente en Traccar")
	}

	// Limpiar todas las cookies relacionadas con la sesión
	cookiesToClear := []string{"JSESSIONID", "traccar-session", "session"}

	for _, cookieName := range cookiesToClear {
		// Limpiar cookie con configuración consistente
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     cookieName,
			Value:    "",
			Path:     "/",
			Domain:   "localhost", // Mismo dominio que en login
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		log.Printf("🧹 [AUTH] Cookie limpiada: %s", cookieName)
	}

	// También limpiar cualquier cookie que el cliente haya enviado
	for _, cookie := range clientCookies {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     cookie.Name,
			Value:    "",
			Path:     "/",
			Domain:   "localhost", // Mismo dominio que en login
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: cookie.HttpOnly,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
		log.Printf("🧹 [AUTH] Cookie del cliente limpiada: %s", cookie.Name)
	}

	log.Printf("✅ [AUTH] Logout completado exitosamente")
	c.JSON(http.StatusOK, gin.H{"message": "Sesión cerrada correctamente"})
}

func (h *AuthHandler) GetSession(c *gin.Context) {
	log.Printf("🔍 [AUTH] Verificando sesión actual")

	// Crear servicio Traccar con las cookies del cliente
	clientCookies := c.Request.Cookies()
	log.Printf("🍪 [AUTH] Cookies recibidas para verificación: %d", len(clientCookies))

	if len(clientCookies) == 0 {
		log.Printf("❌ [AUTH] No hay cookies de sesión")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No hay sesión activa"})
		return
	}

	// Crear servicio con las cookies del cliente
	traccarService := h.traccarService.WithCookies(clientCookies)

	// Verificar sesión con Traccar
	user, err := traccarService.GetSession()
	if err != nil {
		log.Printf("❌ [AUTH] Error verificando sesión: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesión no válida o expirada"})
		return
	}

	log.Printf("✅ [AUTH] Sesión válida para usuario: %s (ID: %d)", user.Name, user.ID)
	c.JSON(http.StatusOK, user)
}
