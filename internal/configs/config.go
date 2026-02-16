package configs

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `yaml:"env" env-default:"local"`
	Postgres PostgressConfig
	TypeDB   string `yaml:"db_type" env:"TYPE_DB" env-default:"postgres"`
}

type PostgressConfig struct {
	Host   string `yaml:"host" env:"POSTGRESS_HOST" env-default:"localhost"`
	Port   string `yaml:"port" env:"POSTGRESS_PORT" env-default:"5432"`
	User   string `yaml:"user" env:"POSTGRES_USER" env-default:"postgress"`
	DBName string `yaml:"db_name" env:"POSTGRESS_DB" env-default:"finance_db"`

	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
}

func LoadConfig() *Config {
	var ConfigPath string = os.Getenv("CONFIG_PATH")

	if ConfigPath == "" {
		ConfigPath = "./configs/config.yaml"
	}

	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		log.Fatalf("Config didnt find %s", ConfigPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(ConfigPath, &cfg); err != nil {
		log.Fatalf("Ошибка чтения конфига: %s", err)
	}

	return &cfg
}
