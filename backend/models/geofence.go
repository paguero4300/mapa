package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Geofence representa una geocerca de Traccar
type Geofence struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Area        string                 `json:"area"`        // Geometría en formato WKT
	CalendarID  int                    `json:"calendarId"`
	Attributes  map[string]interface{} `json:"attributes"`
}

// GeofenceWithGeometry representa una geocerca con geometría parseada para el frontend
type GeofenceWithGeometry struct {
	Geofence
	Geometry GeofenceGeometry `json:"geometry"`
	Type     string           `json:"type"` // "POLYGON", "CIRCLE", etc.
}

// GeofenceGeometry representa la geometría parseada de una geocerca
type GeofenceGeometry struct {
	Type        string      `json:"type"`        // "polygon", "circle"
	Coordinates interface{} `json:"coordinates"` // Coordenadas según el tipo
	Center      *LatLng     `json:"center,omitempty"`      // Para círculos
	Radius      float64     `json:"radius,omitempty"`      // Para círculos (en metros)
}

// LatLng representa un punto de latitud y longitud
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// ParseGeometry parsea el campo Area (WKT) y convierte a geometría para Leaflet
func (g *Geofence) ParseGeometry() (*GeofenceGeometry, error) {
	area := strings.TrimSpace(g.Area)
	if area == "" {
		return nil, fmt.Errorf("área vacía")
	}

	// Detectar tipo de geometría
	areaUpper := strings.ToUpper(area)
	
	if strings.HasPrefix(areaUpper, "POLYGON") {
		return g.parsePolygon(area)
	} else if strings.HasPrefix(areaUpper, "CIRCLE") {
		return g.parseCircle(area)
	} else if strings.HasPrefix(areaUpper, "LINESTRING") {
		return g.parseLineString(area)
	}
	
	return nil, fmt.Errorf("tipo de geometría no soportado: %s", area)
}

// parsePolygon parsea un polígono WKT
func (g *Geofence) parsePolygon(wkt string) (*GeofenceGeometry, error) {
	// Ejemplo: POLYGON ((lat1 lng1, lat2 lng2, lat3 lng3, lat1 lng1))
	re := regexp.MustCompile(`POLYGON\s*\(\s*\((.*?)\)\s*\)`)
	matches := re.FindStringSubmatch(wkt)
	
	if len(matches) < 2 {
		return nil, fmt.Errorf("formato de polígono inválido: %s", wkt)
	}
	
	coordsStr := matches[1]
	coordPairs := strings.Split(coordsStr, ",")
	
	var coordinates [][]float64
	
	for _, pair := range coordPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		
		parts := strings.Fields(pair)
		if len(parts) != 2 {
			continue
		}
		
		lat, err1 := strconv.ParseFloat(parts[0], 64)
		lng, err2 := strconv.ParseFloat(parts[1], 64)
		
		if err1 != nil || err2 != nil {
			continue
		}
		
		// Leaflet espera [lng, lat] para polígonos
		coordinates = append(coordinates, []float64{lng, lat})
	}
	
	if len(coordinates) < 3 {
		return nil, fmt.Errorf("polígono debe tener al menos 3 puntos")
	}
	
	return &GeofenceGeometry{
		Type:        "polygon",
		Coordinates: [][]float64(coordinates),
	}, nil
}

// parseCircle parsea un círculo WKT
func (g *Geofence) parseCircle(wkt string) (*GeofenceGeometry, error) {
	// Ejemplo: CIRCLE (lat lng, radius)
	re := regexp.MustCompile(`CIRCLE\s*\(\s*([-\d.]+)\s+([-\d.]+)\s*,\s*([-\d.]+)\s*\)`)
	matches := re.FindStringSubmatch(wkt)
	
	if len(matches) < 4 {
		return nil, fmt.Errorf("formato de círculo inválido: %s", wkt)
	}
	
	lat, err1 := strconv.ParseFloat(matches[1], 64)
	lng, err2 := strconv.ParseFloat(matches[2], 64)
	radius, err3 := strconv.ParseFloat(matches[3], 64)
	
	if err1 != nil || err2 != nil || err3 != nil {
		return nil, fmt.Errorf("coordenadas de círculo inválidas: %s", wkt)
	}
	
	return &GeofenceGeometry{
		Type: "circle",
		Center: &LatLng{
			Lat: lat,
			Lng: lng,
		},
		Radius: radius,
	}, nil
}

// parseLineString parsea una línea WKT
func (g *Geofence) parseLineString(wkt string) (*GeofenceGeometry, error) {
	// Ejemplo: LINESTRING (lat1 lng1, lat2 lng2, lat3 lng3)
	re := regexp.MustCompile(`LINESTRING\s*\(\s*(.*?)\s*\)`)
	matches := re.FindStringSubmatch(wkt)
	
	if len(matches) < 2 {
		return nil, fmt.Errorf("formato de línea inválido: %s", wkt)
	}
	
	coordsStr := matches[1]
	coordPairs := strings.Split(coordsStr, ",")
	
	var coordinates [][]float64
	
	for _, pair := range coordPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		
		parts := strings.Fields(pair)
		if len(parts) != 2 {
			continue
		}
		
		lat, err1 := strconv.ParseFloat(parts[0], 64)
		lng, err2 := strconv.ParseFloat(parts[1], 64)
		
		if err1 != nil || err2 != nil {
			continue
		}
		
		// Leaflet espera [lng, lat] para líneas
		coordinates = append(coordinates, []float64{lng, lat})
	}
	
	if len(coordinates) < 2 {
		return nil, fmt.Errorf("línea debe tener al menos 2 puntos")
	}
	
	return &GeofenceGeometry{
		Type:        "linestring",
		Coordinates: coordinates,
	}, nil
}

// ToGeofenceWithGeometry convierte una Geofence a GeofenceWithGeometry
func (g *Geofence) ToGeofenceWithGeometry() (*GeofenceWithGeometry, error) {
	geometry, err := g.ParseGeometry()
	if err != nil {
		return nil, err
	}
	
	geofenceType := strings.ToUpper(geometry.Type)
	
	return &GeofenceWithGeometry{
		Geofence: *g,
		Geometry: *geometry,
		Type:     geofenceType,
	}, nil
}

// GeofenceMapData representa los datos de geocercas para el mapa
type GeofenceMapData struct {
	Geofences []*GeofenceWithGeometry `json:"geofences"`
	Count     int                     `json:"count"`
}

// GeofenceUpdateRequest representa una solicitud de actualización de geocerca
type GeofenceUpdateRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Area        string                 `json:"area"`
	CalendarID  int                    `json:"calendarId"`
	Attributes  map[string]interface{} `json:"attributes"`
}

// ToGeofence convierte GeofenceUpdateRequest a Geofence
func (req *GeofenceUpdateRequest) ToGeofence(id int) *Geofence {
	return &Geofence{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Area:        req.Area,
		CalendarID:  req.CalendarID,
		Attributes:  req.Attributes,
	}
}

// IsVisible verifica si la geocerca debe mostrarse en el mapa
func (g *Geofence) IsVisible() bool {
	if g.Attributes == nil {
		return true // Por defecto visible
	}
	
	if visible, exists := g.Attributes["visibleOnMap"]; exists {
		if visibleBool, ok := visible.(bool); ok {
			return visibleBool
		}
	}
	
	return true // Por defecto visible
}

// SetVisible establece la visibilidad de la geocerca
func (g *Geofence) SetVisible(visible bool) {
	if g.Attributes == nil {
		g.Attributes = make(map[string]interface{})
	}
	g.Attributes["visibleOnMap"] = visible
}

// GeofenceVisibilityRequest representa una solicitud de cambio de visibilidad
type GeofenceVisibilityRequest struct {
	Visible bool `json:"visible"`
}