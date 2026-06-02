package config

import "os"

type Config struct {
	Addr   string
	DBPath string
	Agent  AgentConfig
}

type AgentConfig struct {
	Provider      string
	OpenAIAPIKey  string
	OpenAIBaseURL string
	OpenAIModel   string
}

func Load() Config {
	return Config{
		Addr:   getEnv("APP_ADDR", ":8080"),
		DBPath: getEnv("DB_PATH", "./data/doctor.db"),
		Agent: AgentConfig{
			Provider:      getEnv("LLM_PROVIDER", "mock"),
			OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
			OpenAIBaseURL: os.Getenv("OPENAI_BASE_URL"),
			OpenAIModel:   getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
