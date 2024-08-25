package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env     string `envconfig:"ENV" default:"local"`
	Verbose bool   `envconfig:"VERBOSE" default:"false"`
	HTTPServer
	ProxyChecker
	TelegramBot
}

type TelegramBot struct {
	APIToken string `envconfig:"TELEGRAM_API_TOKEN" env-default:""`
}

type HTTPServer struct {
	Address         string        `envconfig:"ADDRESS" default:"localhost:8082"`
	Timeout         time.Duration `envconfig:"REQUEST_TIMEOUT" default:"4s"`
	IdleTimeout     time.Duration `envconfig:"IDLE_TIMEOUT" default:"60s"`
	MaxRequestSize  int64         `envconfig:"MAX_REQUEST_SIZE" default:"1048576"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT"  default:"10s"`
}

type ProxyChecker struct {
	API         string        `envconfig:"API" default:"http://checkip.amazonaws.com"`
	Timeout     time.Duration `envconfig:"CHECKING_TIMEOUT" default:"3600ms"`
	Concurrency uint          `envconfig:"CONCURRENCY" default:"100"`
}

func MustLoad() *Config {
	var cfg Config

	envconfig.MustProcess("", &cfg)

	return &cfg
}
