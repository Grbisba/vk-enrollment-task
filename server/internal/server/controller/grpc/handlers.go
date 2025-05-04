package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Grbisba/vk-enrollment-task/protogen"
)

func (ctrl *Controller) Subscribe(request *pb.SubscribeRequest, g grpc.ServerStreamingServer[pb.Event]) error {
	if request == nil {
		ctrl.log.Error("nil subscribe request")

		return status.Error(codes.InvalidArgument, "request must not be nil")
	}

	sub, err := ctrl.eventBus.Subscribe(request.Key, func(msg interface{}) {
		ctrl.log.Debug(
			"got message for subscriber",
			zap.String("subscriber", request.Key),
			zap.Any("message", msg),
		)

		err := g.Send(&pb.Event{Data: msg.(string)})
		if err != nil {
			ctrl.log.Warn(
				"failed to send event",
				zap.Error(err),
			)

			return
		}
	})
	if err != nil {
		ctrl.log.Error(
			"failed to subscribe to event",
			zap.Error(err),
		)

		return status.Error(codes.Internal, "failed to subscribe to event")
	}

	select {
	case <-g.Context().Done():
		sub.Unsubscribe()
		return status.Error(codes.Canceled, g.Context().Err().Error())
	}
}

func (ctrl *Controller) Publish(_ context.Context, request *pb.PublishRequest) (*emptypb.Empty, error) {
	if request == nil {
		ctrl.log.Error("nil publish request")

		return nil, status.Error(codes.InvalidArgument, "request must not be nil")
	}

	err := ctrl.eventBus.Publish(request.Key, request.Data)
	if err != nil {
		ctrl.log.Error(
			"failed to publish event",
			zap.Error(err),
		)

		return nil, status.Error(codes.Internal, "failed to publish data")
	}

	return &emptypb.Empty{}, nil
}
