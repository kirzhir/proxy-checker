package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env          string `yaml:"env" env:"ENV" env-default:"local"`
	HTTPServer   `yaml:"http_server"`
	ProxyChecker `yaml:"proxy_checker"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type ProxyChecker struct {
	API         string        `yaml:"api" env:"API" env-default:"http://checkip.amazonaws.com"`
	Timeout     time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"5s"`
	Concurrency int           `yaml:"concurrency" env:"CONCURRENCY" env-default:"100"`
}

func MustLoadFile() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Panicf("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Panicf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Panicf("cannot read config: %s", err)
	}

	return &cfg
}

func MustLoadEnv() *Config {

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Panicf("cannot read config: %s", err)
	}

	return &cfg
}
