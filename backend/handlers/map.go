package handlers

import (
	"log"
	"net/http"
	"strconv"
	"traccar-login/models"
	"traccar-login/services"

	"github.com/gin-gonic/gin"
)

type MapHandler struct {
	traccarService *services.TraccarService
}

func NewMapHandler(traccarService *services.TraccarService) *MapHandler {
	return &MapHandler{
		traccarService: traccarService,
	}
}

// MapPage renderiza la página del mapa
func (h *MapHandler) MapPage(c *gin.Context) {
	log.Printf("🗺️ [MAP HANDLER] Renderizando página del mapa")
	
	// Obtener información del usuario del contexto
	userName, _ := c.Get("userName")
	userEmail, _ := c.Get("userEmail")
	userId, _ := c.Get("userId")
	userAdmin, _ := c.Get("userAdmin")
	user, _ := c.Get("user")
	
	c.HTML(http.StatusOK, "map.html", gin.H{
		"Title":     "Mapa en Tiempo Real - Traccar",
		"User":      user,
		"UserId":    userId,
		"UserEmail": userEmail,
		"UserName":  userName,
		"UserAdmin": userAdmin,
	})
	
	log.Printf("✅ [MAP HANDLER] Página del mapa renderizada")
}

// GetMapData obtiene datos para el mapa (posiciones + dispositivos)
func (h *MapHandler) GetMapData(c *gin.Context) {
	log.Printf("🗺️ [MAP HANDLER] Obteniendo datos del mapa")
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [MAP HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Obtener dispositivos y posiciones en paralelo
	devicesChan := make(chan []*models.Device)
	positionsChan := make(chan []*models.Position)
	errorsChan := make(chan error, 2)
	
	// Goroutine para obtener dispositivos
	go func() {
		devices, err := service.GetDevices()
		if err != nil {
			errorsChan <- err
			return
		}
		devicesChan <- devices
	}()
	
	// Goroutine para obtener posiciones
	go func() {
		positions, err := service.GetPositions()
		if err != nil {
			errorsChan <- err
			return
		}
		positionsChan <- positions
	}()
	
	// Esperar resultados
	var devices []*models.Device
	var positions []*models.Position
	var errors []error
	
	for i := 0; i < 2; i++ {
		select {
		case devs := <-devicesChan:
			devices = devs
		case pos := <-positionsChan:
			positions = pos
		case err := <-errorsChan:
			errors = append(errors, err)
		}
	}
	
	// Verificar errores
	if len(errors) > 0 {
		log.Printf("❌ [MAP HANDLER] Errores obteniendo datos: %v", errors)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo datos del mapa"})
		return
	}
	
	// Crear mapa de dispositivos por ID para búsqueda rápida
	deviceMap := make(map[int]*models.Device)
	for _, device := range devices {
		deviceMap[device.ID] = device
	}
	
	// Combinar posiciones con información de dispositivos
	var positionsWithDevices []models.PositionWithDevice
	var totalLat, totalLng float64
	validPositions := 0
	
	for _, position := range positions {
		if device, exists := deviceMap[position.DeviceID]; exists {
			positionWithDevice := models.PositionWithDevice{
				Position:     *position,
				DeviceName:   device.Name,
				DeviceStatus: device.Status,
				DeviceModel:  device.Model,
			}
			positionsWithDevices = append(positionsWithDevices, positionWithDevice)
			
			// Acumular coordenadas para calcular centro
			if position.Valid && position.Latitude != 0 && position.Longitude != 0 {
				totalLat += position.Latitude
				totalLng += position.Longitude
				validPositions++
			}
		}
	}
	
	// Calcular centro del mapa
	center := models.MapCenter{
		Latitude:  -12.0464, // Lima, Perú por defecto
		Longitude: -77.0428,
	}
	
	if validPositions > 0 {
		center.Latitude = totalLat / float64(validPositions)
		center.Longitude = totalLng / float64(validPositions)
	}
	
	// Crear respuesta
	mapData := models.MapData{
		Positions: positionsWithDevices,
		Center:    center,
		Zoom:      13, // Zoom apropiado para ciudad
	}
	
	log.Printf("✅ [MAP HANDLER] Datos del mapa obtenidos: %d posiciones, centro: %.6f,%.6f", 
		len(positionsWithDevices), center.Latitude, center.Longitude)
	
	c.JSON(http.StatusOK, mapData)
}

// GetPositions obtiene solo las posiciones (para actualizaciones en tiempo real)
func (h *MapHandler) GetPositions(c *gin.Context) {
	log.Printf("📍 [MAP HANDLER] Obteniendo posiciones para actualización")
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [MAP HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Obtener posiciones
	positions, err := service.GetPositions()
	if err != nil {
		log.Printf("❌ [MAP HANDLER] Error obteniendo posiciones: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo posiciones"})
		return
	}
	
	log.Printf("✅ [MAP HANDLER] Posiciones obtenidas para actualización: %d", len(positions))
	c.JSON(http.StatusOK, gin.H{
		"positions": positions,
		"timestamp": "now",
	})
}

// GetDevicePosition obtiene la posición de un dispositivo específico
func (h *MapHandler) GetDevicePosition(c *gin.Context) {
	log.Printf("📍 [MAP HANDLER] Obteniendo posición de dispositivo específico")
	
	// Obtener ID del dispositivo
	deviceIDStr := c.Param("id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		log.Printf("❌ [MAP HANDLER] ID de dispositivo inválido: %s", deviceIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de dispositivo inválido"})
		return
	}
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [MAP HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Obtener posición del dispositivo
	position, err := service.GetDevicePosition(deviceID)
	if err != nil {
		log.Printf("❌ [MAP HANDLER] Error obteniendo posición del dispositivo %d: %v", deviceID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Posición no encontrada"})
		return
	}
	
	log.Printf("✅ [MAP HANDLER] Posición del dispositivo %d obtenida", deviceID)
	c.JSON(http.StatusOK, gin.H{
		"position": position,
	})
}

// GetGeofences obtiene las geocercas del usuario
func (h *MapHandler) GetGeofences(c *gin.Context) {
	log.Printf("🔷 [MAP HANDLER] Obteniendo geocercas")
	
	// Obtener el servicio Traccar con cookies del contexto
	traccarService, exists := c.Get("traccarService")
	if !exists {
		log.Printf("❌ [MAP HANDLER] TraccarService no encontrado en contexto")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	service := traccarService.(*services.TraccarService)
	
	// Obtener geocercas con geometría parseada
	geofences, err := service.GetGeofencesWithGeometry()
	if err != nil {
		log.Printf("❌ [MAP HANDLER] Error obteniendo geocercas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo geocercas"})
		return
	}
	
	// Crear respuesta
	geofenceData := models.GeofenceMapData{
		Geofences: geofences,
		Count:     len(geofences),
	}
	
	log.Printf("✅ [MAP HANDLER] Geocercas obtenidas: %d", len(geofences))
	c.JSON(http.StatusOK, geofenceData)
}