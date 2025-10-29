package handlers

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"traccar-login/models"
	"traccar-login/services"

	"github.com/gin-gonic/gin"
)

type GeofencesHandler struct {
	traccarService *services.TraccarService
}

func NewGeofencesHandler(traccarService *services.TraccarService) *GeofencesHandler {
	return &GeofencesHandler{
		traccarService: traccarService,
	}
}

// GeofencesPage renderiza la página de gestión de geocercas
func (h *GeofencesHandler) GeofencesPage(c *gin.Context) {
	log.Printf("🔷 [GEOFENCES HANDLER] Renderizando página de gestión de geocercas")
	
	// Obtener información del usuario del contexto
	userName, _ := c.Get("userName")
	userEmail, _ := c.Get("userEmail")
	userId, _ := c.Get("userId")
	userAdmin, _ := c.Get("userAdmin")
	user, _ := c.Get("user")
	
	c.HTML(http.StatusOK, "geofences.html", gin.H{
		"Title":     "Gestión de Geocercas - Traccar",
		"User":      user,
		"UserId":    userId,
		"UserEmail": userEmail,
		"UserName":  userName,
		"UserAdmin": userAdmin,
	})
	
	log.Printf("✅ [GEOFENCES HANDLER] Página de geocercas renderizada")
}

// GetGeofences obtiene la lista de geocercas
func (h *GeofencesHandler) GetGeofences(c *gin.Context) {
	log.Printf("🔷 [GEOFENCES HANDLER] Obteniendo lista de geocercas")
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [GEOFENCES HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Obtener geocercas
	geofences, err := service.GetGeofences()
	if err != nil {
		log.Printf("❌ [GEOFENCES HANDLER] Error obteniendo geocercas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo geocercas"})
		return
	}
	
	log.Printf("✅ [GEOFENCES HANDLER] Geocercas obtenidas: %d", len(geofences))
	c.JSON(http.StatusOK, gin.H{
		"geofences": geofences,
		"count":     len(geofences),
	})
}

// UpdateGeofence actualiza una geocerca
func (h *GeofencesHandler) UpdateGeofence(c *gin.Context) {
	log.Printf("🔷 [GEOFENCES HANDLER] Actualizando geocerca")
	
	// Obtener ID de la geocerca
	geofenceIDStr := c.Param("id")
	geofenceID, err := strconv.Atoi(geofenceIDStr)
	if err != nil {
		log.Printf("❌ [GEOFENCES HANDLER] ID de geocerca inválido: %s", geofenceIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de geocerca inválido"})
		return
	}
	
	// Obtener datos del request
	var updateReq models.GeofenceUpdateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		log.Printf("❌ [GEOFENCES HANDLER] Error parseando request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [GEOFENCES HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Actualizar geocerca
	updatedGeofence, err := service.UpdateGeofence(geofenceID, &updateReq)
	if err != nil {
		log.Printf("❌ [GEOFENCES HANDLER] Error actualizando geocerca %d: %v", geofenceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando geocerca"})
		return
	}
	
	log.Printf("✅ [GEOFENCES HANDLER] Geocerca %d actualizada", geofenceID)
	c.JSON(http.StatusOK, gin.H{
		"geofence": updatedGeofence,
		"message":  "Geocerca actualizada exitosamente",
	})
}

// DeleteGeofence elimina una geocerca
func (h *GeofencesHandler) DeleteGeofence(c *gin.Context) {
	log.Printf("🔷 [GEOFENCES HANDLER] Eliminando geocerca")
	
	// Obtener ID de la geocerca
	geofenceIDStr := c.Param("id")
	geofenceID, err := strconv.Atoi(geofenceIDStr)
	if err != nil {
		log.Printf("❌ [GEOFENCES HANDLER] ID de geocerca inválido: %s", geofenceIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de geocerca inválido"})
		return
	}
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [GEOFENCES HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Eliminar geocerca
	err = service.DeleteGeofence(geofenceID)
	if err != nil {
		log.Printf("❌ [GEOFENCES HANDLER] Error eliminando geocerca %d: %v", geofenceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando geocerca"})
		return
	}
	
	log.Printf("✅ [GEOFENCES HANDLER] Geocerca %d eliminada", geofenceID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Geocerca eliminada exitosamente",
	})
}

// UpdateGeofenceVisibility actualiza la visibilidad de una geocerca
func (h *GeofencesHandler) UpdateGeofenceVisibility(c *gin.Context) {
	log.Printf("🔷 [GEOFENCES HANDLER] ===== ACTUALIZANDO VISIBILIDAD =====")
	log.Printf("🔷 [GEOFENCES HANDLER] Method: %s", c.Request.Method)
	log.Printf("🔷 [GEOFENCES HANDLER] URL: %s", c.Request.URL.String())
	log.Printf("🔷 [GEOFENCES HANDLER] Headers: %v", c.Request.Header)
	
	// Obtener ID de la geocerca
	geofenceIDStr := c.Param("id")
	geofenceID, err := strconv.Atoi(geofenceIDStr)
	if err != nil {
		log.Printf("❌ [GEOFENCES HANDLER] ID de geocerca inválido: %s", geofenceIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de geocerca inválido"})
		return
	}
	
	// Leer el cuerpo del request para debug
	body, _ := io.ReadAll(c.Request.Body)
	log.Printf("📝 [GEOFENCES HANDLER] Request body: %s", string(body))
	
	// Recrear el reader para que ShouldBindJSON pueda leerlo
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	
	// Obtener datos del request
	var visibilityReq models.GeofenceVisibilityRequest
	if err := c.ShouldBindJSON(&visibilityReq); err != nil {
		log.Printf("❌ [GEOFENCES HANDLER] Error parseando request: %v", err)
		log.Printf("❌ [GEOFENCES HANDLER] Request body era: %s", string(body))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}
	
	log.Printf("✅ [GEOFENCES HANDLER] Request parseado correctamente: visible=%t", visibilityReq.Visible)
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [GEOFENCES HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Actualizar visibilidad
	err = service.UpdateGeofenceVisibility(geofenceID, visibilityReq.Visible)
	if err != nil {
		log.Printf("❌ [GEOFENCES HANDLER] Error actualizando visibilidad geocerca %d: %v", geofenceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando visibilidad"})
		return
	}
	
	log.Printf("✅ [GEOFENCES HANDLER] Visibilidad de geocerca %d actualizada: %t", geofenceID, visibilityReq.Visible)
	c.JSON(http.StatusOK, gin.H{
		"message": "Visibilidad actualizada exitosamente",
		"visible": visibilityReq.Visible,
	})
}