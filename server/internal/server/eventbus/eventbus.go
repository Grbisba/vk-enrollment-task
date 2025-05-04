package eventbus

import (
	"go.uber.org/zap"

	"github.com/Grbisba/vk-enrollment-task/subpub"
)

type EventBus struct {
	subpub.SubPub
	log *zap.Logger
}

func New(log *zap.Logger) *EventBus {
	e := &EventBus{
		SubPub: subpub.NewSubPub(),
		log:    log.Named("eventbus"),
	}
	return e
}
