package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

func LoadConfig() (*Config, error) {
	k := koanf.New(".")

	loadEnvFile(".env.local")

	configFile := "config/config.yaml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configFile = "config/config.yaml.template"
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	processedContent := expandEnvVars(string(content))

	if err := k.Load(rawbytes.Provider([]byte(processedContent)), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil); err != nil {
		return nil, fmt.Errorf("error loading env vars: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

func expandEnvVars(content string) string {
	re := regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		envVar := submatches[1]
		defaultValue := ""
		if len(submatches) > 2 {
			defaultValue = submatches[2]
		}

		if value := os.Getenv(envVar); value != "" {
			return value
		}

		return defaultValue
	})
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
