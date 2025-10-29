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

	log.Printf("✅ [AUTH] Login exitoso, sesión global establecida")

	// IMPORTANTE: Ya no dependemos de cookies del cliente
	// La sesión se mantiene en el servidor backend usando el cookie jar
	log.Printf("�� [AUTH] ===== SESIÓN MANTENIDA EN SERVIDOR =====")

	// Si hay cookies, establecerlas de todas formas (para compatibilidad)
	if len(loginResponse.Cookies) > 0 {
		log.Printf("🍪 [AUTH] Estableciendo %d cookies para compatibilidad", len(loginResponse.Cookies))
		for i, cookie := range loginResponse.Cookies {
			log.Printf("🍪 [AUTH] Cookie %d: %s=%s", i+1, cookie.Name, cookie.Value)

			// Configuración de cookie para desarrollo local
			clientCookie := &http.Cookie{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     "/",
				Domain:   "localhost",
				MaxAge:   86400,
				Secure:   false,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}

			http.SetCookie(c.Writer, clientCookie)
			log.Printf("🍪 [AUTH] Cookie establecida: %s", clientCookie.Name)
		}
	} else {
		log.Printf("�� [AUTH] No hay cookies, usando sesión global del servidor")
	}

	log.Printf("�� [AUTH] ===== SESIÓN GLOBAL ACTIVA =====")

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

	// PRIMERO: Limpiar la sesión global
	services.SetGlobalSession(nil)
	log.Printf("🌐 [AUTH] Sesión global limpiada")

	// SEGUNDO: Intentar cerrar sesión en Traccar si hay sesión activa
	globalSession := services.GetGlobalSession()
	if globalSession != nil {
		err := globalSession.Logout()
		if err != nil {
			log.Printf("⚠️ [AUTH] Error cerrando sesión en Traccar: %v", err)
		} else {
			log.Printf("✅ [AUTH] Sesión cerrada exitosamente en Traccar")
		}
	} else {
		// Si no hay sesión global, intentar con cookies del cliente
		clientCookies := c.Request.Cookies()
		log.Printf("🍪 [AUTH] Cookies del cliente para logout: %d", len(clientCookies))

		if len(clientCookies) > 0 {
			traccarService := h.traccarService.WithCookies(clientCookies)
			err := traccarService.Logout()
			if err != nil {
				log.Printf("⚠️ [AUTH] Error cerrando sesión con cookies: %v", err)
			} else {
				log.Printf("✅ [AUTH] Sesión cerrada exitosamente con cookies")
			}
		}
	}

	// Limpiar todas las cookies relacionadas con la sesión
	cookiesToClear := []string{"JSESSIONID", "traccar-session", "session"}

	for _, cookieName := range cookiesToClear {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     cookieName,
			Value:    "",
			Path:     "/",
			Domain:   "localhost",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
		log.Printf("🧹 [AUTH] Cookie limpiada: %s", cookieName)
	}

	// También limpiar cualquier cookie que el cliente haya enviado
	clientCookies := c.Request.Cookies()
	for _, cookie := range clientCookies {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     cookie.Name,
			Value:    "",
			Path:     "/",
			Domain:   "localhost",
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
