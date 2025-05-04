package client

import (
	"go.uber.org/fx"

	"github.com/Grbisba/vk-enrollment-task/client/internal/config"
)

func main() {
	fx.New(buildOptions()).Run()
}

func buildOptions() fx.Option {
	return fx.Options(
		fx.Provide(
			config.New,
		),
		fx.Invoke(),
	)
}
