package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all environment-driven settings for the app.
type Config struct {
	Port               string
	JWTSecret          string
	RazorpayKeyID      string
	RazorpayKeySecret  string
	RazorpayWebhookKey string
	AppBaseURL         string
}

var App Config

type FactoryLocation struct {
	Latitude     float64
	Longitude    float64
	RadiusMeters float64
}

// Load reads .env (if present) and populates App. Safe to call multiple times.
func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("config: no .env file found, relying on real environment variables")
	}

	App = Config{
		Port:               getEnv("PORT", "8080"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-me"),
		RazorpayKeyID:      getEnv("RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret:  getEnv("RAZORPAY_KEY_SECRET", ""),
		RazorpayWebhookKey: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
		AppBaseURL:         getEnv("APP_BASE_URL", "http://localhost:8080"),
	}

	if App.JWTSecret == "dev-secret-change-me" {
		log.Println("config: WARNING using default JWT_SECRET, set a real one in .env for production")
	}

	return App
}

func LoadFactoryLocation() FactoryLocation {
	lat, err := strconv.ParseFloat(os.Getenv("FACTORY_LATITUDE"), 64)
	if err != nil {
		log.Fatalf("invalid or missing FACTORY_LATITUDE: %v", err)
	}
	lng, err := strconv.ParseFloat(os.Getenv("FACTORY_LONGITUDE"), 64)
	if err != nil {
		log.Fatalf("invalid or missing FACTORY_LONGITUDE: %v", err)
	}

	radius := 100.0 // sensible default
	if r := os.Getenv("FACTORY_RADIUS_METERS"); r != "" {
		if parsed, err := strconv.ParseFloat(r, 64); err == nil {
			radius = parsed
		}
	}

	return FactoryLocation{Latitude: lat, Longitude: lng, RadiusMeters: radius}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
