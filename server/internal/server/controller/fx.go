package controller

import (
	"go.uber.org/fx"
)

func RunControllerFx(lc fx.Lifecycle, ctrl Controller) {
	lc.Append(fx.Hook{
		OnStart: ctrl.Start,
		OnStop:  ctrl.Stop,
	})
}
