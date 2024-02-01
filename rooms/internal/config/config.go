package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `env-default:"local"`
	HttpPort   string `env-default:"8080"`
	HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	var cfg Config

	env := os.Getenv("ENV")
	httpPort := os.Getenv("HTTP_PORT")

	if env == "" || httpPort == "" {
		log.Fatal("required env variables are not set")
	}

	configPath := "configs/config_" + env + ".yml"

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
		log.Fatal(err)
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	} else {
		cfg.Env = env
		cfg.Address += ":" + httpPort
	}

	return &cfg
}
