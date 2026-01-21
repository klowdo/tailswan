package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Port      string
	LogLevel  string
	Tailscale TailscaleConfig
	Swan      SwanConfig
}

type TailscaleConfig struct {
	StateDir    string
	Socket      string
	Hostname    string
	AuthKey     string
	Routes      []string
	SSH         bool
	ExtraArgs   []string
	UseTsnet    bool
	EnableServe bool
}

type SwanConfig struct {
	ConfigPath  string
	AutoStart   bool
	Connections []string
}

func Load() *Config {
	port := getEnv("CONTROL_PORT", "8080")
	logLevel := getEnv("LOG_LEVEL", "info")

	tsStateDir := getEnv("TS_STATE_DIR", "/var/lib/tailscale")
	tsSocket := getEnv("TS_SOCKET", "/var/run/tailscale/tailscaled.sock")
	tsHostname := getEnv("TS_HOSTNAME", "tailswan")
	tsAuthKey := getEnv("TS_AUTHKEY", "")
	tsRoutes := getEnv("TS_ROUTES", "")
	tsSSH := getEnvBool("TS_SSH", false)
	tsExtraArgs := getEnv("TS_EXTRA_ARGS", "")
	useTsnet := getEnvBool("USE_TSNET", false)
	tsEnableServe := getEnvBool("SWAN_TS_SERVE", false)

	swanConfig := getEnv("SWAN_CONFIG", "/etc/swanctl/swanctl.conf")
	swanAutoStart := getEnvBool("SWAN_AUTO_START", false)
	swanConnections := getEnv("SWAN_CONNECTIONS", "")

	cfg := &Config{
		Port:     port,
		LogLevel: logLevel,
		Tailscale: TailscaleConfig{
			StateDir:    tsStateDir,
			Socket:      tsSocket,
			Hostname:    tsHostname,
			AuthKey:     tsAuthKey,
			Routes:      parseCommaSeparated(tsRoutes),
			SSH:         tsSSH,
			ExtraArgs:   strings.Fields(tsExtraArgs),
			UseTsnet:    useTsnet,
			EnableServe: tsEnableServe,
		},
		Swan: SwanConfig{
			ConfigPath:  swanConfig,
			AutoStart:   swanAutoStart,
			Connections: parseCommaSeparated(swanConnections),
		},
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

func (c *Config) Address() string {
	return ":" + c.Port
}

func (c *Config) GetLogLevel() slog.Level {
	switch strings.ToLower(strings.TrimSpace(c.LogLevel)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
