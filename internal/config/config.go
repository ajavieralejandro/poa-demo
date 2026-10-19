// Package config carga la configuración del proceso desde el entorno
// y, si existe, desde un archivo .env. Las variables del sistema tienen prioridad.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort            = "8080"
	defaultDifficulty      = 16
	defaultInitialSupply   = 1000
	defaultMaxBlocks       = 1000
	defaultBlockIntervalMS = 2000
	defaultOrigins         = "http://localhost:5173,http://127.0.0.1:5173,http://localhost:4173,http://127.0.0.1:4173"
)

// Config reúne los parámetros que comparten el servidor HTTP y el motor de minado.
type Config struct {
	Port           string
	Difficulty     int
	InitialSupply  uint32
	MaxBlocks      int
	AllowedOrigins []string
	BlockInterval  time.Duration
}

// ListenAddr devuelve la dirección que escucha el servidor HTTP.
func (c Config) ListenAddr() string {
	port := c.Port
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

// Load lee .env en el directorio de trabajo y después las variables de entorno.
func Load() (Config, error) {
	return load(os.Getenv, ".env")
}

func load(getenv func(string) string, path string) (Config, error) {
	fileValues, err := readDotEnv(path)
	if err != nil {
		return Config{}, err
	}

	lookup := func(key string) string {
		if value := strings.TrimSpace(getenv(key)); value != "" {
			return value
		}
		return strings.TrimSpace(fileValues[key])
	}

	port := lookup("PORT")
	if port == "" {
		port = defaultPort
	}
	port = strings.TrimPrefix(port, ":")
	if _, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("config: PORT %q no es un número", port)
	}
	portNumber, _ := strconv.Atoi(port)
	if portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf("config: PORT %d está fuera de rango", portNumber)
	}

	difficulty, err := parseInt(lookup("DIFFICULTY"), defaultDifficulty, "DIFFICULTY")
	if err != nil {
		return Config{}, err
	}
	if difficulty < 0 || difficulty > 256 {
		return Config{}, fmt.Errorf("config: DIFFICULTY debe estar entre 0 y 256")
	}

	supply, err := parseUint32(lookup("INITIAL_SUPPLY"), defaultInitialSupply, "INITIAL_SUPPLY")
	if err != nil {
		return Config{}, err
	}
	if supply == 0 {
		return Config{}, fmt.Errorf("config: INITIAL_SUPPLY debe ser mayor que cero")
	}

	maxBlocks, err := parseInt(lookup("MAX_BLOCKS"), defaultMaxBlocks, "MAX_BLOCKS")
	if err != nil {
		return Config{}, err
	}
	if maxBlocks < 3 {
		return Config{}, fmt.Errorf("config: MAX_BLOCKS debe ser al menos 3")
	}

	intervalMS, err := parseInt(lookup("BLOCK_INTERVAL_MS"), defaultBlockIntervalMS, "BLOCK_INTERVAL_MS")
	if err != nil {
		return Config{}, err
	}
	if intervalMS < 0 {
		return Config{}, fmt.Errorf("config: BLOCK_INTERVAL_MS no puede ser negativo")
	}

	originsRaw := lookup("ALLOWED_ORIGINS")
	if originsRaw == "" {
		originsRaw = defaultOrigins
	}

	return Config{
		Port:           port,
		Difficulty:     difficulty,
		InitialSupply:  supply,
		MaxBlocks:      maxBlocks,
		AllowedOrigins: splitList(originsRaw),
		BlockInterval:  time.Duration(intervalMS) * time.Millisecond,
	}, nil
}

func parseInt(raw string, fallback int, name string) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("config: %s %q no es un entero", name, raw)
	}
	return value, nil
}

func parseUint32(raw string, fallback uint32, name string) (uint32, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("config: %s %q no es un entero válido", name, raw)
	}
	return uint32(value), nil
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func readDotEnv(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("config: leer %s: %w", path, err)
	}
	return parseDotEnv(string(data)), nil
}

func parseDotEnv(content string) map[string]string {
	values := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			quoted := (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')
			if quoted {
				value = value[1 : len(value)-1]
			}
		}
		if key != "" {
			values[key] = value
		}
	}
	return values
}
