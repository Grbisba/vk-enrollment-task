package grpc

import (
	"context"
	"errors"
	"net"
	"time"

	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/Grbisba/vk-enrollment-task/protogen"

	"github.com/Grbisba/vk-enrollment-task/client/internal/config"
	"github.com/Grbisba/vk-enrollment-task/client/internal/controller"
	"github.com/Grbisba/vk-enrollment-task/client/internal/service"
)

var (
	_ controller.Controller = (*Controller)(nil)
	_ pb.PubSubClient       = (*Controller)(nil)
)

type Controller struct {
	log      *zap.Logger
	server   *grpc.Server
	cfg      *config.Controller
	listener net.Listener
	service  service.Service
}

func newWithConfig(log *zap.Logger, srv service.Service, cfg *config.Controller) (*Controller, error) {
	log = log.Named("grpc")
	ctrl := &Controller{
		log:     log.Named("controller"),
		service: srv,
		cfg:     cfg,
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

func New(log *zap.Logger, srv service.Service, cfg *config.Config) (*Controller, error) {
	return newWithConfig(log, srv, cfg.GRPC)
}

func (ctrl *Controller) createServer(log *zap.Logger) (err error) {
	if ctrl == nil {
		// TODO: move to standalone error
		return errors.New("nil controller")
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
	//pb(ctrl.server, ctrl)

	log.Info("successfully created gRPC server")
	return nil
}

func (ctrl *Controller) createListener() (err error) {
	ctrl.listener, err = net.Listen("tcp", ctrl.cfg.Bind())
	if err != nil {
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
			cancel(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return ctx.Err()
}

func (ctrl *Controller) Stop(_ context.Context) error {
	ctrl.server.GracefulStop()
	return ctrl.listener.Close()
}
