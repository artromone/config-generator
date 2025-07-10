// config/loader.go
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

func LoadConfig() (*Config, error) {
	// 1. Загружаем базовую конфигурацию из YAML
	configFile := "config/config.yaml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configFile = "config/config.yaml.template"
	}

	if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	// 2. Загружаем переменные окружения из .env.local
	loadEnvFile(".env.local")

	// 3. Переопределяем системными переменными окружения
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.Replace(strings.ToLower(s), "_", ".", -1)
	}), nil); err != nil {
		return nil, fmt.Errorf("error loading env vars: %w", err)
	}

	// 4. Unmarshaling в структуру
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// 5. Валидация
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func loadEnvFile(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				os.Setenv(key, value)
			}
		}
	}
}
