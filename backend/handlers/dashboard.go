package handlers

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) Dashboard(c *gin.Context) {
	log.Printf("🏠 [DASHBOARD HANDLER] Procesando request para dashboard")
	
	// Obtener información del usuario del contexto (seteada por middleware)
	user, exists := c.Get("user")
	if !exists {
		log.Printf("❌ [DASHBOARD HANDLER] Usuario no encontrado en contexto")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}
	
	// Obtener datos adicionales del contexto
	userId, _ := c.Get("userId")
	userEmail, _ := c.Get("userEmail")
	userName, _ := c.Get("userName")
	userAdmin, _ := c.Get("userAdmin")
	
	log.Printf("✅ [DASHBOARD HANDLER] Renderizando dashboard para usuario: %s (ID: %v)", userName, userId)
	log.Printf("📋 [DASHBOARD HANDLER] Datos del template: Title=Dashboard, UserName=%s, UserEmail=%s, UserAdmin=%v", userName, userEmail, userAdmin)
	
	// IMPORTANTE: Verificar que estamos renderizando el template correcto
	log.Printf("📄 [DASHBOARD HANDLER] Renderizando template: dashboard.html")
	
	templateData := gin.H{
		"Title":     "Dashboard - Traccar",
		"User":      user,
		"UserId":    userId,
		"UserEmail": userEmail,
		"UserName":  userName,
		"UserAdmin": userAdmin,
	}
	
	log.Printf("📋 [DASHBOARD HANDLER] Template data: %+v", templateData)
	
	c.HTML(http.StatusOK, "dashboard.html", templateData)
	
	log.Printf("✅ [DASHBOARD HANDLER] Dashboard renderizado exitosamente con template dashboard.html")
}