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
	assert.Equal(uint(100), cfg.ProxyChecker.Concurrency)
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
