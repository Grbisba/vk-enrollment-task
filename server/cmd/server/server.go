package main

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/Grbisba/vk-enrollment-task/server/internal/server/config"
	"github.com/Grbisba/vk-enrollment-task/server/internal/server/controller/grpc"
	"github.com/Grbisba/vk-enrollment-task/server/internal/server/eventbus"
	"github.com/Grbisba/vk-enrollment-task/server/pkg/logger"

	"github.com/Grbisba/vk-enrollment-task/server/internal/server/controller"
)

func main() {
	fx.New(buildOptions()).Run()
}

func buildOptions() fx.Option {
	return fx.Options(
		fx.WithLogger(zapLogger),
		fx.Provide(
			newLogger,
			config.New,
			eventbus.New,

			fx.Annotate(grpc.New, fx.As(new(controller.Controller))),
		),
		fx.Invoke(
			controller.RunControllerFx,
		),
	)
}

func newLogger() *zap.Logger {
	l, _ := logger.NewProduction()
	return l.Named("client")
}

func zapLogger(log *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{
		Logger: log.Named("fx"),
	}
}
