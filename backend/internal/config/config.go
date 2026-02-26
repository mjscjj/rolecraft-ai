package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	Env             string
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	OpenAIKey       string
	MilvusAddr      string
	AnythingLLMURL  string // AnythingLLM API URL
	AnythingLLMKey  string // AnythingLLM API Key
}

// Load 加载配置
func Load() *Config {
	return &Config{
		Env:            getEnv("ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://rolecraft:rolecraft@localhost:5432/rolecraft?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		OpenAIKey:      getEnv("OPENAI_API_KEY", ""),
		MilvusAddr:     getEnv("MILVUS_ADDR", "localhost:19530"),
		AnythingLLMURL: getEnv("ANYTHINGLLM_URL", "http://150.109.21.115:3001"),
		AnythingLLMKey: getEnv("ANYTHINGLLM_KEY", "sk-WaUmgZsMxgeHOpp8SJxK1rmVQxiwfiDJ"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}