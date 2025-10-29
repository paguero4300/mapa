package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TraccarURL  string
	ServerPort  string
	JWTSecret   string
	Environment string
	IsLocalhost bool
}

func LoadConfig() (*Config, error) {
	// Intentar cargar .env.local primero (para desarrollo)
	err := godotenv.Load("../.env.local")
	if err == nil {
		log.Printf("✅ .env.local cargado desde directorio raíz (modo desarrollo)")
	} else {
		// No se encontró .env.local, cargar .env para producción
		err = godotenv.Load("../.env")
		if err == nil {
			log.Printf("✅ .env cargado desde directorio raíz (modo producción)")
		} else {
			// Intentar cargar .env local
			err = godotenv.Load()
			if err == nil {
				log.Printf("✅ .env cargado desde directorio local")
			} else {
				log.Printf("No se pudo cargar .env local, usando variables del sistema")
			}
		}
	}

	config := &Config{
		TraccarURL:  getEnv("TRACCAR_URL", "https://demo.traccar.org/api"),
		ServerPort:  getEnv("PORT", getEnv("SERVER_PORT", "8080")), // Soporta ambas variables
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		Environment: getEnv("ENV", "development"),
	}

	// Determinar si estamos en localhost
	config.IsLocalhost = config.Environment == "development" ||
		getEnv("LOCALHOST", "false") == "true"

	// Mostrar configuración cargada (sin mostrar el JWT secret por seguridad)
	log.Printf("🔧 Configuración cargada:")
	log.Printf("   TraccarURL: %s", config.TraccarURL)
	log.Printf("   ServerPort: %s", config.ServerPort)
	log.Printf("   Environment: %s", config.Environment)
	log.Printf("   IsLocalhost: %t", config.IsLocalhost)
	log.Printf("   JWTSecret: [OCULTO]")

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
