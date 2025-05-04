package grpc

import (
	"context"
	"net"
	"time"

	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/Grbisba/vk-enrollment-task/protogen"
	"github.com/Grbisba/vk-enrollment-task/server/internal/server/config"
	"github.com/Grbisba/vk-enrollment-task/server/internal/server/controller"
	"github.com/Grbisba/vk-enrollment-task/server/internal/server/eventbus"
	"github.com/Grbisba/vk-enrollment-task/subpub"
)

var (
	_ controller.Controller = (*Controller)(nil)
	_ pb.PubSubServer       = (*Controller)(nil)
)

type Controller struct {
	pb.UnimplementedPubSubServer
	log      *zap.Logger
	server   *grpc.Server
	cfg      *config.Controller
	listener net.Listener
	eventBus subpub.SubPub
}

func newWithConfig(log *zap.Logger, cfg *config.Controller, eBus *eventbus.EventBus) (*Controller, error) {
	log = log.Named("grpc")
	ctrl := &Controller{
		log:      log.Named("controller"),
		cfg:      cfg,
		eventBus: eBus,
	}

	err := multierr.Combine(
		ctrl.createListener(),
		ctrl.createServer(log),
	)
	if err != nil {
		return nil, err
	}

	return ctrl, nil
}

func New(log *zap.Logger, cfg *config.Config, eBus *eventbus.EventBus) (*Controller, error) {
	return newWithConfig(log, cfg.Controller, eBus)
}

func (ctrl *Controller) createServer(log *zap.Logger) (err error) {
	if ctrl == nil {
		return errNilController
	}

	if log == nil {
		log = zap.L()
	}

	log = log.Named("internal")

	grpcZap.ReplaceGrpcLoggerV2(log)

	log.Info("creating gRPC server")

	var opts = []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpcRecovery.UnaryServerInterceptor(),
		),
	}

	ctrl.server = grpc.NewServer(opts...)
	pb.RegisterPubSubServer(ctrl.server, ctrl)

	log.Info("successfully created gRPC server")
	return nil
}

func (ctrl *Controller) createListener() (err error) {
	ctrl.listener, err = net.Listen("tcp", ctrl.cfg.Bind())
	if err != nil {
		ctrl.log.Error(
			"failed to create listener",
			zap.Error(err),
		)

		return err
	}

	ctrl.log.Info("created listener")
	return nil
}

func (ctrl *Controller) Start(ctx context.Context) error {
	var cancel context.CancelCauseFunc
	ctrl.log.Info("Start just called")

	ctx, cancel = context.WithCancelCause(ctx)

	go func() {
		err := ctrl.server.Serve(ctrl.listener)
		if err != nil {
			ctrl.log.Error(
				"failed to start gRPC server",
				zap.Error(err),
			)

			cancel(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return ctx.Err()
}

func (ctrl *Controller) Stop(ctx context.Context) error {
	ctrl.server.GracefulStop()
	err := multierr.Combine(
		ctrl.eventBus.Close(ctx),
		ctrl.listener.Close(),
	)
	if err != nil {
		ctrl.log.Error(
			"failed to graceful stop gRPC server",
			zap.Error(err),
		)

		return err
	}

	return nil
}
