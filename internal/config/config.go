package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env          string `yaml:"env" env:"ENV" env-default:"local"`
	Verbose      bool   `yaml:"verbose" env:"VERBOSE" env-default:"false"`
	HTTPServer   `yaml:"http_server"`
	ProxyChecker `yaml:"proxy_checker"`
	TelegramBot  `yaml:"telegram_bot"`
}

type TelegramBot struct {
	APIToken string `yaml:"api_token" env:"TELEGRAM_API_TOKEN" env-default:""`
}

type HTTPServer struct {
	Address        string        `yaml:"address" env-default:"localhost:8082"`
	Timeout        time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout    time.Duration `yaml:"idle_timeout" env-default:"60s"`
	MaxRequestSize int64         `yaml:"max_request_size" env-default:"1048576"`
}

type ProxyChecker struct {
	API         string        `yaml:"api" env:"API" env-default:"http://checkip.amazonaws.com"`
	Timeout     time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"3600ms"`
	Concurrency int           `yaml:"concurrency" env:"CONCURRENCY" env-default:"100"`
}

func MustLoad() *Config {
	var cfg Config
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Panicf("cannot read config: %s", err)
		}

		return &cfg
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Panicf("config file does not exist: %s", configPath)
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Panicf("cannot read config: %s", err)
	}

	return &cfg
}
