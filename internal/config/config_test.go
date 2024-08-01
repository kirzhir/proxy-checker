package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTempConfigFile(content string) (string, error) {
	dir, err := filepath.Abs("testdata")
	if err != nil {
		return "", err
	}

	tmpfile, err := os.CreateTemp(dir, "config-*.yaml")

	if err != nil {
		return "", err
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		tmpfile.Close()
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}

func TestMustLoadFile(t *testing.T) {
	configContent := `
env: test
http_server:
  address: "localhost:9090"
  timeout: 10s
  idle_timeout: 120s
proxy_checker:
  api: "http://example.com"
  timeout: 10s
  concurrency: 50
`
	configPath, err := createTempConfigFile(configContent)
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	defer os.Remove(configPath)

	os.Setenv("CONFIG_PATH", configPath)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := MustLoadFile()

	assert.Equal(t, "test", cfg.Env)
	assert.Equal(t, "localhost:9090", cfg.HTTPServer.Address)
	assert.Equal(t, 10*time.Second, cfg.HTTPServer.Timeout)
	assert.Equal(t, 120*time.Second, cfg.HTTPServer.IdleTimeout)
	assert.Equal(t, "http://example.com", cfg.ProxyChecker.API)
	assert.Equal(t, 10*time.Second, cfg.ProxyChecker.Timeout)
	assert.Equal(t, 50, cfg.ProxyChecker.Concurrency)
}

func TestMustLoadFileConfigPathNotSet(t *testing.T) {
	os.Unsetenv("CONFIG_PATH")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic but did not get one")
		}
	}()

	MustLoadFile()

}

func TestMustLoadFileConfigPathDoesNotExist(t *testing.T) {
	os.Setenv("CONFIG_PATH", "/nonexistent/path")
	defer os.Unsetenv("CONFIG_PATH")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic but did not get one")
		}

	}()

	MustLoadFile()

}

func TestMustLoadEnv(t *testing.T) {
	os.Setenv("ENV", "test")
	os.Setenv("API", "http://example.com")
	os.Setenv("TIMEOUT", "10s")
	os.Setenv("CONCURRENCY", "50")

	defer func() {
		os.Unsetenv("ENV")
		os.Unsetenv("API")
		os.Unsetenv("TIMEOUT")
		os.Unsetenv("CONCURRENCY")
	}()

	cfg := MustLoadEnv()

	assert.Equal(t, "test", cfg.Env)
	assert.Equal(t, "http://example.com", cfg.ProxyChecker.API)
	assert.Equal(t, 10*time.Second, cfg.ProxyChecker.Timeout)
	assert.Equal(t, 50, cfg.ProxyChecker.Concurrency)
}

func TestMustLoadEnvDefaultValues(t *testing.T) {
	cfg := MustLoadEnv()

	assert.Equal(t, "local", cfg.Env)
	assert.Equal(t, "localhost:8080", cfg.HTTPServer.Address)
	assert.Equal(t, 4*time.Second, cfg.HTTPServer.Timeout)
	assert.Equal(t, 60*time.Second, cfg.HTTPServer.IdleTimeout)
	assert.Equal(t, "http://checkip.amazonaws.com", cfg.ProxyChecker.API)
	assert.Equal(t, 5*time.Second, cfg.ProxyChecker.Timeout)
	assert.Equal(t, 100, cfg.ProxyChecker.Concurrency)
}
