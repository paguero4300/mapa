package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	TraccarURL string
	ServerPort string
	JWTSecret  string
}

func LoadConfig() (*Config, error) {
	// Cargar variables de entorno desde el directorio raíz
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("No se pudo cargar .env desde ../, intentando local...")
		// Si no encuentra en ../, intentar en el directorio actual
		err = godotenv.Load()
		if err != nil {
			log.Printf("No se pudo cargar .env local, usando variables del sistema")
		} else {
			log.Printf("✅ .env cargado desde directorio local")
		}
	} else {
		log.Printf("✅ .env cargado desde directorio raíz")
	}
	
	config := &Config{
		TraccarURL: getEnv("TRACCAR_URL", "https://demo.traccar.org/api"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
	}
	
	// Mostrar configuración cargada (sin mostrar el JWT secret por seguridad)
	log.Printf("🔧 Configuración cargada:")
	log.Printf("   TraccarURL: %s", config.TraccarURL)
	log.Printf("   ServerPort: %s", config.ServerPort)
	log.Printf("   JWTSecret: [OCULTO]")
	
	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}