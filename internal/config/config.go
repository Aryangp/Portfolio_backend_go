package config

import (
	"os"
	"strings"
)

// Config holds all configuration values for the application
type Config struct {
	Port              string
	AppEnv            string
	AllowedOrigins    []string
	MongoURI          string
	MongoDBName       string
	MongoProjectsColl string
	GeminiAPIKey      string
	GeminiModel       string
	GitHubUsername    string
	GitHubToken       string
}

// Load loads configuration from environment variables (or .env file) with fallback defaults
func Load() *Config {
	loadDotEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb+srv://AryanTest:Aryangp05@testingmongocluster.f5oeqdc.mongodb.net/?appName=testingMongoCluster"
	}

	mongoDBName := os.Getenv("MONGO_DB_NAME")
	if mongoDBName == "" {
		mongoDBName = "portfolio_db"
	}

	mongoProjectsColl := os.Getenv("MONGO_PROJECTS_COLL")
	if mongoProjectsColl == "" {
		mongoProjectsColl = "projects"
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-1.5-flash"
	}

	githubUsername := os.Getenv("GITHUB_USERNAME")
	if githubUsername == "" {
		githubUsername = "Aryangp"
	}

	githubToken := os.Getenv("GITHUB_TOKEN")

	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv == "" || allowedOriginsEnv == "*" {
		allowedOrigins = []string{"*"}
	} else {
		for _, origin := range strings.Split(allowedOriginsEnv, ",") {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	return &Config{
		Port:              port,
		AppEnv:            appEnv,
		AllowedOrigins:    allowedOrigins,
		MongoURI:          mongoURI,
		MongoDBName:       mongoDBName,
		MongoProjectsColl: mongoProjectsColl,
		GeminiAPIKey:      geminiAPIKey,
		GeminiModel:       geminiModel,
		GitHubUsername:    githubUsername,
		GitHubToken:       githubToken,
	}
}

// loadDotEnv reads key=value pairs from a local .env file and sets them in os.Environ if not already set
func loadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			_ = os.Setenv(key, val)
		}
	}
}
