package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
	"traccar-login/models"
)

type TraccarService struct {
	BaseURL string
	Client  *http.Client
}

func NewTraccarService(baseURL string) *TraccarService {
	// Crear cookie jar para manejar sesiones automáticamente
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Printf("Error creando cookie jar: %v", err)
	}

	return &TraccarService{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
			// Importante: permitir redirecciones y manejar cookies correctamente
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Permitir redirecciones automáticamente
				return nil
			},
		},
	}
}

func (s *TraccarService) Login(email, password string) (*models.LoginResponse, error) {
	log.Printf("🔐 [TRACCAR] Iniciando login para: %s", email)

	// IMPORTANTE: Traccar requiere application/x-www-form-urlencoded según la API
	formData := url.Values{}
	formData.Set("email", email)
	formData.Set("password", password)

	log.Printf("📡 [TRACCAR] Enviando request a: %s/session", s.BaseURL)
	log.Printf("📡 [TRACCAR] Datos: email=%s, password=[HIDDEN]", email)

	// Crear request con form data
	req, err := http.NewRequest("POST", s.BaseURL+"/session", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	// Headers correctos para Traccar
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	log.Printf("📡 [TRACCAR] Enviando request: %s %s", req.Method, req.URL.String())
	log.Printf("📋 [TRACCAR] Headers: %v", req.Header)
	log.Printf("📝 [TRACCAR] Body: %s", formData.Encode())

	// Ejecutar request
	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error ejecutando request: %v", err)
		return nil, fmt.Errorf("error conectando con Traccar: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Status Code: %d", resp.StatusCode)
	log.Printf("📋 [TRACCAR] Response Headers: %v", resp.Header)

	// Verificar status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [TRACCAR] Login fallido - Status: %d", resp.StatusCode)
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("credenciales inválidas")
		}
		return nil, fmt.Errorf("error del servidor Traccar: %d", resp.StatusCode)
	}

	// Parsear respuesta JSON
	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Printf("❌ [TRACCAR] Error parseando JSON: %v", err)
		return nil, fmt.Errorf("error parseando respuesta del servidor: %v", err)
	}

	log.Printf("✅ [TRACCAR] Usuario autenticado: %s (ID: %d)", user.Name, user.ID)

	// Extraer cookies de la respuesta - MÉTODO MEJORADO
	cookies := resp.Cookies()

	// Si no hay cookies con el método automático, intentar extraer manualmente
	if len(cookies) == 0 {
		log.Printf("🔍 [TRACCAR] No se encontraron cookies automáticamente, extrayendo manualmente...")
		cookies = extractCookiesManual(resp)
	}

	log.Printf("🍪 [TRACCAR] Cookies recibidas: %d", len(cookies))

	for i, cookie := range cookies {
		log.Printf("🍪 [TRACCAR] Cookie %d: %s=%s (Domain: %s, Path: %s, HttpOnly: %v, Secure: %v)",
			i+1, cookie.Name, cookie.Value, cookie.Domain, cookie.Path, cookie.HttpOnly, cookie.Secure)
	}

	// Crear respuesta
	response := &models.LoginResponse{
		User:    &user,
		Cookies: cookies,
	}

	log.Printf("✅ [TRACCAR] Login completado exitosamente")
	return response, nil
}

func (s *TraccarService) GetSession() (*models.User, error) {
	sessionURL := fmt.Sprintf("%s/session", s.BaseURL)

	log.Printf("🔍 [TRACCAR] Verificando sesión en: %s", sessionURL)

	req, err := http.NewRequest("GET", sessionURL, nil)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error creando request: %v", err)
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	// Headers para verificación de sesión
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Navegar-App/1.0")

	// Log de cookies que se envían
	if s.Client.Jar != nil {
		if parsedURL, err := url.Parse(sessionURL); err == nil {
			cookies := s.Client.Jar.Cookies(parsedURL)
			log.Printf("🍪 [TRACCAR] Enviando %d cookies para verificación", len(cookies))
			for _, cookie := range cookies {
				log.Printf("🍪 [TRACCAR] Cookie enviada: %s=%s", cookie.Name, cookie.Value)
			}
		}
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error verificando sesión: %v", err)
		return nil, fmt.Errorf("error verificando sesión: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Session check status: %s", resp.Status)

	// Leer el cuerpo para debug
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error leyendo response body: %v", err)
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	log.Printf("📄 [TRACCAR] Session response body: %s", string(body))

	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("🚫 [TRACCAR] Sesión no válida o expirada (401)")
		return nil, fmt.Errorf("sesión no válida o expirada")
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("🚫 [TRACCAR] Sesión no encontrada (404)")
		return nil, fmt.Errorf("sesión no encontrada")
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [TRACCAR] Error verificando sesión: %d", resp.StatusCode)
		return nil, fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		log.Printf("❌ [TRACCAR] Error parseando usuario: %v", err)
		return nil, fmt.Errorf("error parseando respuesta: %w", err)
	}

	log.Printf("✅ [TRACCAR] Sesión válida para usuario: ID=%d, Name=%s, Email=%s",
		user.ID, user.Name, user.Email)

	return &user, nil
}

func (s *TraccarService) Logout() error {
	sessionURL := fmt.Sprintf("%s/session", s.BaseURL)

	log.Printf("🚪 [TRACCAR] Cerrando sesión en: %s", sessionURL)

	req, err := http.NewRequest("DELETE", sessionURL, nil)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error creando request de logout: %v", err)
		return fmt.Errorf("error creando request: %w", err)
	}

	// Headers para logout
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Navegar-App/1.0")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error cerrando sesión: %v", err)
		return fmt.Errorf("error cerrando sesión: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Logout status: %s", resp.Status)

	// Leer respuesta para debug
	body, _ := io.ReadAll(resp.Body)
	log.Printf("📄 [TRACCAR] Logout response: %s", string(body))

	// El logout puede devolver 204 (No Content) o 200
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("⚠️ [TRACCAR] Logout con status inesperado: %d", resp.StatusCode)
		// No fallar por esto, puede que la sesión ya haya expirado
	}

	log.Printf("✅ [TRACCAR] Sesión cerrada exitosamente")
	return nil
}

// ===== DISPOSITIVOS =====

// GetDevices obtiene la lista de dispositivos del usuario
func (s *TraccarService) GetDevices() ([]*models.Device, error) {
	devicesURL := fmt.Sprintf("%s/devices", s.BaseURL)
	log.Printf("📱 [TRACCAR] Obteniendo dispositivos desde: %s", devicesURL)

	req, err := http.NewRequest("GET", devicesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error obteniendo dispositivos: %v", err)
		return nil, fmt.Errorf("error obteniendo dispositivos: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Devices response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [TRACCAR] Error obteniendo dispositivos: %d", resp.StatusCode)
		return nil, fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	var devices []*models.Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		log.Printf("❌ [TRACCAR] Error parseando dispositivos: %v", err)
		return nil, fmt.Errorf("error parseando respuesta: %v", err)
	}

	log.Printf("✅ [TRACCAR] Dispositivos obtenidos: %d", len(devices))
	return devices, nil
}

// CreateDevice crea un nuevo dispositivo
func (s *TraccarService) CreateDevice(deviceReq *models.DeviceRequest) (*models.Device, error) {
	devicesURL := fmt.Sprintf("%s/devices", s.BaseURL)
	log.Printf("🆕 [TRACCAR] Creando dispositivo: %s", deviceReq.Name)

	jsonData, err := json.Marshal(deviceReq)
	if err != nil {
		return nil, fmt.Errorf("error serializando dispositivo: %v", err)
	}

	req, err := http.NewRequest("POST", devicesURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error creando dispositivo: %v", err)
		return nil, fmt.Errorf("error creando dispositivo: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Create device response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [TRACCAR] Error creando dispositivo: %d", resp.StatusCode)
		return nil, fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	var device models.Device
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		log.Printf("❌ [TRACCAR] Error parseando dispositivo creado: %v", err)
		return nil, fmt.Errorf("error parseando respuesta: %v", err)
	}

	log.Printf("✅ [TRACCAR] Dispositivo creado: ID=%d, Name=%s", device.ID, device.Name)
	return &device, nil
}

// UpdateDevice actualiza un dispositivo existente
func (s *TraccarService) UpdateDevice(deviceID int, deviceReq *models.DeviceRequest) (*models.Device, error) {
	deviceURL := fmt.Sprintf("%s/devices/%d", s.BaseURL, deviceID)
	log.Printf("✏️ [TRACCAR] Actualizando dispositivo ID=%d: %s", deviceID, deviceReq.Name)

	jsonData, err := json.Marshal(deviceReq)
	if err != nil {
		return nil, fmt.Errorf("error serializando dispositivo: %v", err)
	}

	req, err := http.NewRequest("PUT", deviceURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error actualizando dispositivo: %v", err)
		return nil, fmt.Errorf("error actualizando dispositivo: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Update device response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [TRACCAR] Error actualizando dispositivo: %d", resp.StatusCode)
		return nil, fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	var device models.Device
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		log.Printf("❌ [TRACCAR] Error parseando dispositivo actualizado: %v", err)
		return nil, fmt.Errorf("error parseando respuesta: %v", err)
	}

	log.Printf("✅ [TRACCAR] Dispositivo actualizado: ID=%d, Name=%s", device.ID, device.Name)
	return &device, nil
}

// DeleteDevice elimina un dispositivo
func (s *TraccarService) DeleteDevice(deviceID int) error {
	deviceURL := fmt.Sprintf("%s/devices/%d", s.BaseURL, deviceID)
	log.Printf("🗑️ [TRACCAR] Eliminando dispositivo ID=%d", deviceID)

	req, err := http.NewRequest("DELETE", deviceURL, nil)
	if err != nil {
		return fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error eliminando dispositivo: %v", err)
		return fmt.Errorf("error eliminando dispositivo: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Delete device response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusNoContent {
		log.Printf("❌ [TRACCAR] Error eliminando dispositivo: %d", resp.StatusCode)
		return fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	log.Printf("✅ [TRACCAR] Dispositivo eliminado exitosamente: ID=%d", deviceID)
	return nil
}

// ===== POSICIONES =====

// GetPositions obtiene las últimas posiciones conocidas de todos los dispositivos
func (s *TraccarService) GetPositions() ([]*models.Position, error) {
	positionsURL := fmt.Sprintf("%s/positions", s.BaseURL)
	log.Printf("📍 [TRACCAR] Obteniendo posiciones desde: %s", positionsURL)

	req, err := http.NewRequest("GET", positionsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error obteniendo posiciones: %v", err)
		return nil, fmt.Errorf("error obteniendo posiciones: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Positions response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [TRACCAR] Error obteniendo posiciones: %d", resp.StatusCode)
		return nil, fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	var positions []*models.Position
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		log.Printf("❌ [TRACCAR] Error parseando posiciones: %v", err)
		return nil, fmt.Errorf("error parseando respuesta: %v", err)
	}

	log.Printf("✅ [TRACCAR] Posiciones obtenidas: %d", len(positions))
	return positions, nil
}

// GetDevicePosition obtiene la última posición de un dispositivo específico
func (s *TraccarService) GetDevicePosition(deviceID int) (*models.Position, error) {
	positionsURL := fmt.Sprintf("%s/positions?deviceId=%d", s.BaseURL, deviceID)
	log.Printf("📍 [TRACCAR] Obteniendo posición del dispositivo %d desde: %s", deviceID, positionsURL)

	req, err := http.NewRequest("GET", positionsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error obteniendo posición del dispositivo: %v", err)
		return nil, fmt.Errorf("error obteniendo posición: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Device position response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [TRACCAR] Error obteniendo posición del dispositivo: %d", resp.StatusCode)
		return nil, fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	var positions []*models.Position
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		log.Printf("❌ [TRACCAR] Error parseando posición del dispositivo: %v", err)
		return nil, fmt.Errorf("error parseando respuesta: %v", err)
	}

	if len(positions) == 0 {
		log.Printf("⚠️ [TRACCAR] No se encontró posición para el dispositivo %d", deviceID)
		return nil, fmt.Errorf("no se encontró posición para el dispositivo")
	}

	log.Printf("✅ [TRACCAR] Posición del dispositivo %d obtenida", deviceID)
	return positions[0], nil
}

// ===== GEOCERCAS =====

// GetGeofences obtiene todas las geocercas del usuario
func (s *TraccarService) GetGeofences() ([]*models.Geofence, error) {
	geofencesURL := fmt.Sprintf("%s/geofences", s.BaseURL)
	log.Printf("🔷 [TRACCAR] Obteniendo geocercas desde: %s", geofencesURL)

	req, err := http.NewRequest("GET", geofencesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error obteniendo geocercas: %v", err)
		return nil, fmt.Errorf("error obteniendo geocercas: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Geofences response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [TRACCAR] Error obteniendo geocercas: %d", resp.StatusCode)
		return nil, fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	var geofences []*models.Geofence
	if err := json.NewDecoder(resp.Body).Decode(&geofences); err != nil {
		log.Printf("❌ [TRACCAR] Error parseando geocercas: %v", err)
		return nil, fmt.Errorf("error parseando respuesta: %v", err)
	}

	log.Printf("✅ [TRACCAR] Geocercas obtenidas: %d", len(geofences))
	return geofences, nil
}

// GetGeofencesWithGeometry obtiene geocercas con geometría parseada para el frontend
func (s *TraccarService) GetGeofencesWithGeometry() ([]*models.GeofenceWithGeometry, error) {
	geofences, err := s.GetGeofences()
	if err != nil {
		return nil, err
	}

	var result []*models.GeofenceWithGeometry

	for _, geofence := range geofences {
		geofenceWithGeometry, err := geofence.ToGeofenceWithGeometry()
		if err != nil {
			// Log error pero continúa con otras geocercas
			log.Printf("⚠️ [TRACCAR] Error parsing geofence %d (%s): %v", geofence.ID, geofence.Name, err)
			continue
		}

		result = append(result, geofenceWithGeometry)
	}

	log.Printf("✅ [TRACCAR] Geocercas con geometría parseadas: %d de %d", len(result), len(geofences))
	return result, nil
}

// UpdateGeofence actualiza una geocerca existente
func (s *TraccarService) UpdateGeofence(geofenceID int, updateReq *models.GeofenceUpdateRequest) (*models.Geofence, error) {
	geofenceURL := fmt.Sprintf("%s/geofences/%d", s.BaseURL, geofenceID)
	log.Printf("🔷 [TRACCAR] Actualizando geocerca ID=%d: %s", geofenceID, updateReq.Name)

	// Convertir a Geofence completo
	geofence := updateReq.ToGeofence(geofenceID)

	jsonData, err := json.Marshal(geofence)
	if err != nil {
		return nil, fmt.Errorf("error serializando geocerca: %v", err)
	}

	req, err := http.NewRequest("PUT", geofenceURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error actualizando geocerca: %v", err)
		return nil, fmt.Errorf("error actualizando geocerca: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Update geofence response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [TRACCAR] Error actualizando geocerca: %d", resp.StatusCode)
		return nil, fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	var updatedGeofence models.Geofence
	if err := json.NewDecoder(resp.Body).Decode(&updatedGeofence); err != nil {
		log.Printf("❌ [TRACCAR] Error parseando geocerca actualizada: %v", err)
		return nil, fmt.Errorf("error parseando respuesta: %v", err)
	}

	log.Printf("✅ [TRACCAR] Geocerca actualizada: ID=%d, Name=%s", updatedGeofence.ID, updatedGeofence.Name)
	return &updatedGeofence, nil
}

// DeleteGeofence elimina una geocerca
func (s *TraccarService) DeleteGeofence(geofenceID int) error {
	geofenceURL := fmt.Sprintf("%s/geofences/%d", s.BaseURL, geofenceID)
	log.Printf("🔷 [TRACCAR] Eliminando geocerca ID=%d", geofenceID)

	req, err := http.NewRequest("DELETE", geofenceURL, nil)
	if err != nil {
		return fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("❌ [TRACCAR] Error eliminando geocerca: %v", err)
		return fmt.Errorf("error eliminando geocerca: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📊 [TRACCAR] Delete geofence response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusNoContent {
		log.Printf("❌ [TRACCAR] Error eliminando geocerca: %d", resp.StatusCode)
		return fmt.Errorf("error del servidor: %d", resp.StatusCode)
	}

	log.Printf("✅ [TRACCAR] Geocerca eliminada exitosamente: ID=%d", geofenceID)
	return nil
}

// UpdateGeofenceVisibility actualiza solo la visibilidad de una geocerca
func (s *TraccarService) UpdateGeofenceVisibility(geofenceID int, visible bool) error {
	log.Printf("🔷 [TRACCAR] Actualizando visibilidad geocerca ID=%d: visible=%t", geofenceID, visible)

	// Primero obtener la geocerca actual
	geofences, err := s.GetGeofences()
	if err != nil {
		return fmt.Errorf("error obteniendo geocercas: %v", err)
	}

	// Buscar la geocerca específica
	var targetGeofence *models.Geofence
	for _, geofence := range geofences {
		if geofence.ID == geofenceID {
			targetGeofence = geofence
			break
		}
	}

	if targetGeofence == nil {
		return fmt.Errorf("geocerca con ID %d no encontrada", geofenceID)
	}

	// Actualizar visibilidad
	targetGeofence.SetVisible(visible)

	// Crear request de actualización
	updateReq := &models.GeofenceUpdateRequest{
		Name:        targetGeofence.Name,
		Description: targetGeofence.Description,
		Area:        targetGeofence.Area,
		CalendarID:  targetGeofence.CalendarID,
		Attributes:  targetGeofence.Attributes,
	}

	// Actualizar en Traccar
	_, err = s.UpdateGeofence(geofenceID, updateReq)
	if err != nil {
		return fmt.Errorf("error actualizando geocerca: %v", err)
	}

	log.Printf("✅ [TRACCAR] Visibilidad de geocerca %d actualizada: %t", geofenceID, visible)
	return nil
}

// extractCookiesManual extrae cookies manualmente del header Set-Cookie
func extractCookiesManual(resp *http.Response) []*http.Cookie {
	var cookies []*http.Cookie

	// Obtener todos los headers Set-Cookie
	setCookieHeaders := resp.Header.Values("Set-Cookie")
	log.Printf("🔍 [TRACCAR] Headers Set-Cookie encontrados: %d", len(setCookieHeaders))

	for i, header := range setCookieHeaders {
		log.Printf("🔍 [TRACCAR] Set-Cookie %d: %s", i+1, header)

		// Parsear manualmente el header
		parts := strings.SplitN(header, ";", 2)
		if len(parts) == 0 {
			continue
		}

		// Extraer nombre y valor
		nameValue := strings.TrimSpace(parts[0])
		if nameValue == "" {
			continue
		}

		nameValueParts := strings.SplitN(nameValue, "=", 2)
		if len(nameValueParts) != 2 {
			continue
		}

		name := strings.TrimSpace(nameValueParts[0])
		value := strings.TrimSpace(nameValueParts[1])

		// Crear cookie
		cookie := &http.Cookie{
			Name:  name,
			Value: value,
			Path:  "/", // Default path
		}

		// Parsear atributos adicionales si existen
		if len(parts) > 1 {
			attributes := strings.Split(parts[1], ";")
			for _, attr := range attributes {
				attr = strings.TrimSpace(attr)
				if strings.HasPrefix(attr, "Path=") {
					cookie.Path = strings.TrimPrefix(attr, "Path=")
				} else if strings.HasPrefix(attr, "Domain=") {
					cookie.Domain = strings.TrimPrefix(attr, "Domain=")
				} else if attr == "HttpOnly" {
					cookie.HttpOnly = true
				} else if attr == "Secure" {
					cookie.Secure = true
				} else if strings.HasPrefix(attr, "SameSite=") {
					sameSite := strings.TrimPrefix(attr, "SameSite=")
					switch strings.ToLower(sameSite) {
					case "strict":
						cookie.SameSite = http.SameSiteStrictMode
					case "lax":
						cookie.SameSite = http.SameSiteLaxMode
					case "none":
						cookie.SameSite = http.SameSiteNoneMode
					}
				}
			}
		}

		cookies = append(cookies, cookie)
		log.Printf("✅ [TRACCAR] Cookie extraída manualmente: %s=%s (Path: %s, Domain: %s)",
			cookie.Name, cookie.Value, cookie.Path, cookie.Domain)
	}

	return cookies
}

// Helper para crear un nuevo servicio con cookies específicas
func (s *TraccarService) WithCookies(cookies []*http.Cookie) *TraccarService {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Printf("Error creando cookie jar: %v", err)
		return s
	}

	// Crear URL base para las cookies
	baseURL, err := url.Parse(s.BaseURL)
	if err != nil {
		log.Printf("Error parseando baseURL: %v", err)
		return s
	}

	// Establecer cookies en el jar
	jar.SetCookies(baseURL, cookies)

	return &TraccarService{
		BaseURL: s.BaseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}
