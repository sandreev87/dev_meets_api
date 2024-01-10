package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `env-default:"local"`
	Postgresql `yaml:"postgresql"`
	HTTPServer `yaml:"http_server"`
}

type Postgresql struct {
	Host           string `yaml:"host" env-default:"0.0.0.0"`
	Port           int    `yaml:"port" env-default:"5432"`
	User           string `env-default:""`
	Password       string `env-default:""`
	DB             string `env-default:""`
	MigrationsPath string `yaml:"migrations_path" env-default:"/migrations"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	var cfg Config

	env := os.Getenv("ENV")
	pgUser := os.Getenv("POSTGRES_USER")
	pgPassword := os.Getenv("POSTGRES_PASSWORD")
	pgDB := os.Getenv("POSTGRES_DB")

	if env == "" || pgUser == "" || pgPassword == "" || pgDB == "" {
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
		cfg.Postgresql.User = pgUser
		cfg.Postgresql.Password = pgPassword
		cfg.Postgresql.DB = pgDB
	}

	return &cfg
}
