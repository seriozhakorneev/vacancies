package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config модель данных необходимых для авторизации
type Config struct {
	Domain          string `yaml:"domain"`
	Keyword         string `yaml:"keyword"`
	AuthorizeTarget string `yaml:"authorize_target"`
}

// NewConfig возвращает необходиммые данные
// авторизации снятые с конфига
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("cleanenv.ReadConfig: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("cleanenv.ReadEnv: %w", err)
	}

	return cfg, nil
}
