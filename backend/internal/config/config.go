package config

import (
	"os"
	"strings"
)

// Config 应用配置
type Config struct {
	Env             string
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	OpenAIKey       string
	OpenRouterURL   string // OpenRouter API URL
	OpenRouterKey   string // OpenRouter API Key
	OpenRouterModel string // OpenRouter 默认模型
	MilvusAddr      string
	AnythingLLMURL  string // AnythingLLM API URL
	AnythingLLMKey  string // AnythingLLM API Key
}

// Load 加载配置
func Load() *Config {
	anythingURL := firstNonEmpty(
		os.Getenv("ANYTHINGLLM_BASE_URL"),
		os.Getenv("ANYTHINGLLM_URL"),
	)
	anythingKey := firstNonEmpty(
		os.Getenv("ANYTHINGLLM_API_KEY"),
		os.Getenv("ANYTHINGLLM_KEY"),
	)

	return &Config{
		Env:             getEnv("ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "sqlite://./rolecraft.db"), // 默认 SQLite，零配置
		RedisURL:        getEnv("REDIS_URL", ""),                           // 可选，空则禁用
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		OpenAIKey:       getEnv("OPENAI_API_KEY", ""), // 可选，使用 Mock AI
		OpenRouterURL:   getEnv("OPENROUTER_URL", "https://openrouter.ai/api/v1"),
		OpenRouterKey:   getEnv("OPENROUTER_KEY", ""),
		OpenRouterModel: getEnv("OPENROUTER_MODEL", "google/gemini-3-flash-preview"),
		MilvusAddr:      getEnv("MILVUS_ADDR", ""), // 可选，空则禁用
		AnythingLLMURL:  normalizeAnythingLLMRootURL(anythingURL),
		AnythingLLMKey:  anythingKey,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func normalizeAnythingLLMRootURL(raw string) string {
	base := strings.TrimSpace(raw)
	if base == "" {
		return ""
	}
	base = strings.TrimRight(base, "/")
	base = strings.TrimSuffix(base, "/api/v1")
	base = strings.TrimSuffix(base, "/api")
	return strings.TrimRight(base, "/")
}
