package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMustLoad_Defaults(t *testing.T) {
	os.Clearenv()

	cfg := MustLoad()

	assert := assert.New(t)
	assert.Equal("local", cfg.Env)
	assert.False(cfg.Verbose)
	assert.Equal("localhost:8082", cfg.HTTPServer.Address)
	assert.Equal(4*time.Second, cfg.HTTPServer.Timeout)
	assert.Equal(60*time.Second, cfg.HTTPServer.IdleTimeout)
	assert.Equal(int64(1048576), cfg.HTTPServer.MaxRequestSize)
	assert.Equal(10*time.Second, cfg.HTTPServer.ShutdownTimeout)
	assert.Equal("http://checkip.amazonaws.com", cfg.ProxyChecker.API)
	assert.Equal(3600*time.Millisecond, cfg.ProxyChecker.Timeout)
	assert.Equal(100, cfg.ProxyChecker.Concurrency)
	assert.Equal("", cfg.TelegramBot.APIToken)
}

func TestMustLoad_EnvVariables(t *testing.T) {
	os.Clearenv()

	os.Setenv("ENV", "production")
	os.Setenv("VERBOSE", "true")
	os.Setenv("TELEGRAM_API_TOKEN", "test-token")
	os.Setenv("SHUTDOWN_TIMEOUT", "15s")
	os.Setenv("CHECKING_TIMEOUT", "5s")

	cfg := MustLoad()

	assert := assert.New(t)
	assert.Equal("production", cfg.Env)
	assert.True(cfg.Verbose)
	assert.Equal(15*time.Second, cfg.HTTPServer.ShutdownTimeout)
	assert.Equal(5*time.Second, cfg.ProxyChecker.Timeout)
	assert.Equal("test-token", cfg.TelegramBot.APIToken)
}

func TestMustLoad_ConfigFile(t *testing.T) {
	os.Clearenv()

	os.Setenv("CONFIG_PATH", "testdata/config.yaml")
	cfg := MustLoad()

	assert := assert.New(t)
	assert.Equal("test", cfg.Env)
	assert.Equal("127.0.0.1:9090", cfg.HTTPServer.Address)
	assert.Equal(5*time.Second, cfg.HTTPServer.Timeout)
	assert.Equal("http://example.com", cfg.ProxyChecker.API)
	assert.Equal("file-token", cfg.TelegramBot.APIToken)
}

func TestMustLoad_ConfigFileNotExist(t *testing.T) {
	os.Setenv("CONFIG_PATH", "non-existent.yaml")

	assert.PanicsWithValue(t, "config file does not exist: non-existent.yaml", func() {
		MustLoad()
	})
}

func TestMustLoad_InvalidConfigFile(t *testing.T) {
	invalidConfigContent := `invalid yaml content`

	file, err := os.CreateTemp("", "invalid_config.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(file.Name())

	if _, err := file.Write([]byte(invalidConfigContent)); err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	file.Close()

	os.Setenv("CONFIG_PATH", file.Name())

	assert.Panics(t, func() {
		MustLoad()
	})
}
