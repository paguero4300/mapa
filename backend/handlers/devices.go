package handlers

import (
	"log"
	"net/http"
	"strconv"
	"traccar-login/models"
	"traccar-login/services"

	"github.com/gin-gonic/gin"
)

type DevicesHandler struct {
	traccarService *services.TraccarService
}

func NewDevicesHandler(traccarService *services.TraccarService) *DevicesHandler {
	return &DevicesHandler{
		traccarService: traccarService,
	}
}

// DevicesPage renderiza la página de dispositivos
func (h *DevicesHandler) DevicesPage(c *gin.Context) {
	log.Printf("📱 [DEVICES HANDLER] Renderizando página de dispositivos")
	
	// Obtener información del usuario del contexto
	userName, _ := c.Get("userName")
	userEmail, _ := c.Get("userEmail")
	userId, _ := c.Get("userId")
	userAdmin, _ := c.Get("userAdmin")
	user, _ := c.Get("user")
	
	c.HTML(http.StatusOK, "devices.html", gin.H{
		"Title":     "Dispositivos - Traccar",
		"User":      user,
		"UserId":    userId,
		"UserEmail": userEmail,
		"UserName":  userName,
		"UserAdmin": userAdmin,
	})
	
	log.Printf("✅ [DEVICES HANDLER] Página de dispositivos renderizada")
}

// GetDevices obtiene la lista de dispositivos
func (h *DevicesHandler) GetDevices(c *gin.Context) {
	log.Printf("📱 [DEVICES HANDLER] Obteniendo lista de dispositivos")
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [DEVICES HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Obtener dispositivos desde Traccar
	devices, err := service.GetDevices()
	if err != nil {
		log.Printf("❌ [DEVICES HANDLER] Error obteniendo dispositivos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo dispositivos"})
		return
	}
	
	log.Printf("✅ [DEVICES HANDLER] Dispositivos obtenidos: %d", len(devices))
	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
		"count":   len(devices),
	})
}

// CreateDevice crea un nuevo dispositivo
func (h *DevicesHandler) CreateDevice(c *gin.Context) {
	log.Printf("🆕 [DEVICES HANDLER] Creando nuevo dispositivo")
	
	var deviceReq models.DeviceRequest
	if err := c.ShouldBindJSON(&deviceReq); err != nil {
		log.Printf("❌ [DEVICES HANDLER] Error validando datos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}
	
	log.Printf("📝 [DEVICES HANDLER] Datos del dispositivo: Name=%s, UniqueID=%s", deviceReq.Name, deviceReq.UniqueID)
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [DEVICES HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Crear dispositivo en Traccar
	device, err := service.CreateDevice(&deviceReq)
	if err != nil {
		log.Printf("❌ [DEVICES HANDLER] Error creando dispositivo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando dispositivo: " + err.Error()})
		return
	}
	
	log.Printf("✅ [DEVICES HANDLER] Dispositivo creado: ID=%d, Name=%s", device.ID, device.Name)
	c.JSON(http.StatusOK, gin.H{
		"device":  device,
		"message": "Dispositivo creado exitosamente",
	})
}

// UpdateDevice actualiza un dispositivo existente
func (h *DevicesHandler) UpdateDevice(c *gin.Context) {
	log.Printf("✏️ [DEVICES HANDLER] Actualizando dispositivo")
	
	// Obtener ID del dispositivo
	deviceIDStr := c.Param("id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		log.Printf("❌ [DEVICES HANDLER] ID de dispositivo inválido: %s", deviceIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de dispositivo inválido"})
		return
	}
	
	var deviceReq models.DeviceRequest
	if err := c.ShouldBindJSON(&deviceReq); err != nil {
		log.Printf("❌ [DEVICES HANDLER] Error validando datos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}
	
	log.Printf("📝 [DEVICES HANDLER] Actualizando dispositivo ID=%d: Name=%s", deviceID, deviceReq.Name)
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [DEVICES HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Actualizar dispositivo en Traccar
	device, err := service.UpdateDevice(deviceID, &deviceReq)
	if err != nil {
		log.Printf("❌ [DEVICES HANDLER] Error actualizando dispositivo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando dispositivo: " + err.Error()})
		return
	}
	
	log.Printf("✅ [DEVICES HANDLER] Dispositivo actualizado: ID=%d, Name=%s", device.ID, device.Name)
	c.JSON(http.StatusOK, gin.H{
		"device":  device,
		"message": "Dispositivo actualizado exitosamente",
	})
}

// DeleteDevice elimina un dispositivo
func (h *DevicesHandler) DeleteDevice(c *gin.Context) {
	log.Printf("🗑️ [DEVICES HANDLER] Eliminando dispositivo")
	
	// Obtener ID del dispositivo
	deviceIDStr := c.Param("id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		log.Printf("❌ [DEVICES HANDLER] ID de dispositivo inválido: %s", deviceIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de dispositivo inválido"})
		return
	}
	
	log.Printf("🗑️ [DEVICES HANDLER] Eliminando dispositivo ID=%d", deviceID)
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [DEVICES HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Eliminar dispositivo en Traccar
	err = service.DeleteDevice(deviceID)
	if err != nil {
		log.Printf("❌ [DEVICES HANDLER] Error eliminando dispositivo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando dispositivo: " + err.Error()})
		return
	}
	
	log.Printf("✅ [DEVICES HANDLER] Dispositivo eliminado exitosamente: ID=%d", deviceID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Dispositivo eliminado exitosamente",
	})
}