package config

import (
	"context"
	"os"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var configPath = os.Getenv("CONFIG_FILE_PATH")

type Config struct {
	Controller *Controller `config:"controller" toml:"controller" yaml:"controller" json:"controller"`
}

func New() (*Config, error) {
	cfg := &Config{}

	l := confita.NewLoader(
		file.NewBackend(configPath),
	)

	err := l.Load(context.Background(), cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error while loading config")
	}

	zap.NewNop().Named("config").Info("loaded config", zap.Any("config", cfg))

	return cfg, nil
}
