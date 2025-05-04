package production

import (
	"go.uber.org/zap"

	"github.com/Grbisba/vk-enrollment-task/client/internal/config"
)

type Service struct {
	log *zap.Logger
	cfg *config.Config
}

func New(log *zap.Logger, cfg *config.Config) (*Service, error) {
	return &Service{log: log, cfg: cfg}, nil
}
