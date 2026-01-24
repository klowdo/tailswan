package config

import (
	"log/slog"
	"testing"
)

func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single value",
			input:    "192.168.1.0/24",
			expected: []string{"192.168.1.0/24"},
		},
		{
			name:     "multiple values",
			input:    "192.168.1.0/24,10.0.0.0/8,172.16.0.0/12",
			expected: []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"},
		},
		{
			name:     "values with spaces",
			input:    " 192.168.1.0/24 , 10.0.0.0/8 , 172.16.0.0/12 ",
			expected: []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"},
		},
		{
			name:     "empty values filtered",
			input:    "192.168.1.0/24,,10.0.0.0/8",
			expected: []string{"192.168.1.0/24", "10.0.0.0/8"},
		},
		{
			name:     "only commas",
			input:    ",,,",
			expected: []string{},
		},
		{
			name:     "whitespace only values filtered",
			input:    "value1,   ,value2",
			expected: []string{"value1", "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommaSeparated(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("length mismatch: expected %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("index %d: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
		setEnv       bool
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_GETENV_UNSET",
			defaultValue: "default_value",
			setEnv:       false,
			expected:     "default_value",
		},
		{
			name:         "returns env value when set",
			key:          "TEST_GETENV_SET",
			defaultValue: "default_value",
			envValue:     "custom_value",
			setEnv:       true,
			expected:     "custom_value",
		},
		{
			name:         "returns default for empty env value",
			key:          "TEST_GETENV_EMPTY",
			defaultValue: "default_value",
			envValue:     "",
			setEnv:       true,
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue bool
		setEnv       bool
		expected     bool
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_GETENVBOOL_UNSET",
			defaultValue: false,
			setEnv:       false,
			expected:     false,
		},
		{
			name:         "returns true default when env not set",
			key:          "TEST_GETENVBOOL_UNSET_TRUE",
			defaultValue: true,
			setEnv:       false,
			expected:     true,
		},
		{
			name:         "parses true",
			key:          "TEST_GETENVBOOL_TRUE",
			defaultValue: false,
			envValue:     "true",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "parses 1",
			key:          "TEST_GETENVBOOL_ONE",
			defaultValue: false,
			envValue:     "1",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "parses yes",
			key:          "TEST_GETENVBOOL_YES",
			defaultValue: false,
			envValue:     "yes",
			setEnv:       true,
			expected:     true,
		},
		{
			name:         "parses false",
			key:          "TEST_GETENVBOOL_FALSE",
			defaultValue: true,
			envValue:     "false",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "parses 0",
			key:          "TEST_GETENVBOOL_ZERO",
			defaultValue: true,
			envValue:     "0",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "parses no",
			key:          "TEST_GETENVBOOL_NO",
			defaultValue: true,
			envValue:     "no",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "invalid value returns false",
			key:          "TEST_GETENVBOOL_INVALID",
			defaultValue: true,
			envValue:     "invalid",
			setEnv:       true,
			expected:     false,
		},
		{
			name:         "empty value returns default",
			key:          "TEST_GETENVBOOL_EMPTY",
			defaultValue: true,
			envValue:     "",
			setEnv:       true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			result := getEnvBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConfigAddress(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected string
	}{
		{
			name:     "default port",
			port:     "8080",
			expected: ":8080",
		},
		{
			name:     "custom port",
			port:     "3000",
			expected: ":3000",
		},
		{
			name:     "empty port",
			port:     "",
			expected: ":",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Port: tt.port}
			result := cfg.Address()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestConfigGetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected slog.Level
	}{
		{
			name:     "debug level",
			logLevel: "debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "info level",
			logLevel: "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "warn level",
			logLevel: "warn",
			expected: slog.LevelWarn,
		},
		{
			name:     "warning level",
			logLevel: "warning",
			expected: slog.LevelWarn,
		},
		{
			name:     "error level",
			logLevel: "error",
			expected: slog.LevelError,
		},
		{
			name:     "uppercase DEBUG",
			logLevel: "DEBUG",
			expected: slog.LevelDebug,
		},
		{
			name:     "mixed case Info",
			logLevel: "Info",
			expected: slog.LevelInfo,
		},
		{
			name:     "with leading whitespace",
			logLevel: "  debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "with trailing whitespace",
			logLevel: "debug  ",
			expected: slog.LevelDebug,
		},
		{
			name:     "unknown level defaults to info",
			logLevel: "unknown",
			expected: slog.LevelInfo,
		},
		{
			name:     "empty string defaults to info",
			logLevel: "",
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{LogLevel: tt.logLevel}
			result := cfg.GetLogLevel()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

//nolint:gocyclo // test function with many assertions
func TestLoad(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		envVars := []string{
			"CONTROL_PORT", "LOG_LEVEL",
			"TS_STATE_DIR", "TS_SOCKET", "TS_HOSTNAME", "TS_AUTHKEY",
			"TS_ROUTES", "TS_SSH", "TS_EXTRA_ARGS", "USE_TSNET", "SWAN_TS_SERVE",
			"SWAN_CONFIG", "SWAN_AUTO_START", "SWAN_CONNECTIONS",
		}
		for _, v := range envVars {
			t.Setenv(v, "")
		}

		cfg := Load()

		if cfg.Port != "8080" {
			t.Errorf("expected Port %q, got %q", "8080", cfg.Port)
		}
		if cfg.LogLevel != "info" {
			t.Errorf("expected LogLevel %q, got %q", "info", cfg.LogLevel)
		}
		if cfg.Tailscale.StateDir != "/var/lib/tailscale" {
			t.Errorf("expected StateDir %q, got %q", "/var/lib/tailscale", cfg.Tailscale.StateDir)
		}
		if cfg.Tailscale.Socket != "/var/run/tailscale/tailscaled.sock" {
			t.Errorf("expected Socket %q, got %q", "/var/run/tailscale/tailscaled.sock", cfg.Tailscale.Socket)
		}
		if cfg.Tailscale.Hostname != "tailswan" {
			t.Errorf("expected Hostname %q, got %q", "tailswan", cfg.Tailscale.Hostname)
		}
		if cfg.Tailscale.AuthKey != "" {
			t.Errorf("expected AuthKey %q, got %q", "", cfg.Tailscale.AuthKey)
		}
		if len(cfg.Tailscale.Routes) != 0 {
			t.Errorf("expected empty Routes, got %v", cfg.Tailscale.Routes)
		}
		if cfg.Tailscale.SSH != false {
			t.Errorf("expected SSH %v, got %v", false, cfg.Tailscale.SSH)
		}
		if len(cfg.Tailscale.ExtraArgs) != 0 {
			t.Errorf("expected empty ExtraArgs, got %v", cfg.Tailscale.ExtraArgs)
		}
		if cfg.Tailscale.UseTsnet != false {
			t.Errorf("expected UseTsnet %v, got %v", false, cfg.Tailscale.UseTsnet)
		}
		if cfg.Tailscale.EnableServe != false {
			t.Errorf("expected EnableServe %v, got %v", false, cfg.Tailscale.EnableServe)
		}
		if cfg.Swan.ConfigPath != "/etc/swanctl/swanctl.conf" {
			t.Errorf("expected ConfigPath %q, got %q", "/etc/swanctl/swanctl.conf", cfg.Swan.ConfigPath)
		}
		if cfg.Swan.AutoStart != false {
			t.Errorf("expected AutoStart %v, got %v", false, cfg.Swan.AutoStart)
		}
		if len(cfg.Swan.Connections) != 0 {
			t.Errorf("expected empty Connections, got %v", cfg.Swan.Connections)
		}
	})

	t.Run("custom values from environment", func(t *testing.T) {
		t.Setenv("CONTROL_PORT", "9090")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("TS_STATE_DIR", "/custom/state")
		t.Setenv("TS_SOCKET", "/custom/socket.sock")
		t.Setenv("TS_HOSTNAME", "custom-host")
		t.Setenv("TS_AUTHKEY", "tskey-auth-xxx")
		t.Setenv("TS_ROUTES", "192.168.1.0/24,10.0.0.0/8")
		t.Setenv("TS_SSH", "true")
		t.Setenv("TS_EXTRA_ARGS", "--advertise-exit-node --accept-routes")
		t.Setenv("USE_TSNET", "1")
		t.Setenv("SWAN_TS_SERVE", "yes")
		t.Setenv("SWAN_CONFIG", "/custom/swanctl.conf")
		t.Setenv("SWAN_AUTO_START", "true")
		t.Setenv("SWAN_CONNECTIONS", "vpn1,vpn2,vpn3")

		cfg := Load()

		if cfg.Port != "9090" {
			t.Errorf("expected Port %q, got %q", "9090", cfg.Port)
		}
		if cfg.LogLevel != "debug" {
			t.Errorf("expected LogLevel %q, got %q", "debug", cfg.LogLevel)
		}
		if cfg.Tailscale.StateDir != "/custom/state" {
			t.Errorf("expected StateDir %q, got %q", "/custom/state", cfg.Tailscale.StateDir)
		}
		if cfg.Tailscale.Socket != "/custom/socket.sock" {
			t.Errorf("expected Socket %q, got %q", "/custom/socket.sock", cfg.Tailscale.Socket)
		}
		if cfg.Tailscale.Hostname != "custom-host" {
			t.Errorf("expected Hostname %q, got %q", "custom-host", cfg.Tailscale.Hostname)
		}
		if cfg.Tailscale.AuthKey != "tskey-auth-xxx" {
			t.Errorf("expected AuthKey %q, got %q", "tskey-auth-xxx", cfg.Tailscale.AuthKey)
		}
		expectedRoutes := []string{"192.168.1.0/24", "10.0.0.0/8"}
		if len(cfg.Tailscale.Routes) != len(expectedRoutes) {
			t.Errorf("expected Routes %v, got %v", expectedRoutes, cfg.Tailscale.Routes)
		}
		for i, r := range cfg.Tailscale.Routes {
			if r != expectedRoutes[i] {
				t.Errorf("expected Routes[%d] %q, got %q", i, expectedRoutes[i], r)
			}
		}
		if cfg.Tailscale.SSH != true {
			t.Errorf("expected SSH %v, got %v", true, cfg.Tailscale.SSH)
		}
		expectedExtraArgs := []string{"--advertise-exit-node", "--accept-routes"}
		if len(cfg.Tailscale.ExtraArgs) != len(expectedExtraArgs) {
			t.Errorf("expected ExtraArgs %v, got %v", expectedExtraArgs, cfg.Tailscale.ExtraArgs)
		}
		for i, a := range cfg.Tailscale.ExtraArgs {
			if a != expectedExtraArgs[i] {
				t.Errorf("expected ExtraArgs[%d] %q, got %q", i, expectedExtraArgs[i], a)
			}
		}
		if cfg.Tailscale.UseTsnet != true {
			t.Errorf("expected UseTsnet %v, got %v", true, cfg.Tailscale.UseTsnet)
		}
		if cfg.Tailscale.EnableServe != true {
			t.Errorf("expected EnableServe %v, got %v", true, cfg.Tailscale.EnableServe)
		}
		if cfg.Swan.ConfigPath != "/custom/swanctl.conf" {
			t.Errorf("expected ConfigPath %q, got %q", "/custom/swanctl.conf", cfg.Swan.ConfigPath)
		}
		if cfg.Swan.AutoStart != true {
			t.Errorf("expected AutoStart %v, got %v", true, cfg.Swan.AutoStart)
		}
		expectedConnections := []string{"vpn1", "vpn2", "vpn3"}
		if len(cfg.Swan.Connections) != len(expectedConnections) {
			t.Errorf("expected Connections %v, got %v", expectedConnections, cfg.Swan.Connections)
		}
		for i, c := range cfg.Swan.Connections {
			if c != expectedConnections[i] {
				t.Errorf("expected Connections[%d] %q, got %q", i, expectedConnections[i], c)
			}
		}
	})
}

func TestTailscaleConfig(t *testing.T) {
	t.Run("struct fields", func(t *testing.T) {
		tc := TailscaleConfig{
			StateDir:    "/state",
			Socket:      "/socket.sock",
			Hostname:    "host",
			AuthKey:     "key",
			Routes:      []string{"10.0.0.0/8"},
			SSH:         true,
			ExtraArgs:   []string{"--arg1"},
			UseTsnet:    true,
			EnableServe: true,
		}

		if tc.StateDir != "/state" {
			t.Errorf("expected StateDir %q, got %q", "/state", tc.StateDir)
		}
		if tc.Socket != "/socket.sock" {
			t.Errorf("expected Socket %q, got %q", "/socket.sock", tc.Socket)
		}
		if tc.Hostname != "host" {
			t.Errorf("expected Hostname %q, got %q", "host", tc.Hostname)
		}
		if tc.AuthKey != "key" {
			t.Errorf("expected AuthKey %q, got %q", "key", tc.AuthKey)
		}
		if len(tc.Routes) != 1 || tc.Routes[0] != "10.0.0.0/8" {
			t.Errorf("expected Routes %v, got %v", []string{"10.0.0.0/8"}, tc.Routes)
		}
		if !tc.SSH {
			t.Error("expected SSH to be true")
		}
		if len(tc.ExtraArgs) != 1 || tc.ExtraArgs[0] != "--arg1" {
			t.Errorf("expected ExtraArgs %v, got %v", []string{"--arg1"}, tc.ExtraArgs)
		}
		if !tc.UseTsnet {
			t.Error("expected UseTsnet to be true")
		}
		if !tc.EnableServe {
			t.Error("expected EnableServe to be true")
		}
	})
}

func TestSwanConfig(t *testing.T) {
	t.Run("struct fields", func(t *testing.T) {
		sc := SwanConfig{
			ConfigPath:  "/etc/swanctl.conf",
			AutoStart:   true,
			Connections: []string{"conn1", "conn2"},
		}

		if sc.ConfigPath != "/etc/swanctl.conf" {
			t.Errorf("expected ConfigPath %q, got %q", "/etc/swanctl.conf", sc.ConfigPath)
		}
		if !sc.AutoStart {
			t.Error("expected AutoStart to be true")
		}
		if len(sc.Connections) != 2 {
			t.Errorf("expected 2 Connections, got %d", len(sc.Connections))
		}
		if sc.Connections[0] != "conn1" || sc.Connections[1] != "conn2" {
			t.Errorf("expected Connections %v, got %v", []string{"conn1", "conn2"}, sc.Connections)
		}
	})
}
