package rpc

import (
	"context"
	"detect/master/conf"
	"detect/master/model"
	"detect/master/receiver"
	pb "detect/protos"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

const (
	RECEIVER_WORKER_COUNT = 10
)

type RpcServer struct {
	engine *xorm.Engine
	config *conf.MasterConfig
	logger *logrus.Logger
	rs     []receiver.Receiver
}

func NewRpcServer(engine *xorm.Engine, config *conf.MasterConfig, logger *logrus.Logger) *RpcServer {
	rs := receiver.NewReceivers(config, logger, engine)
	rpcServer := &RpcServer{
		engine: engine,
		config: config,
		logger: logger,
		rs:     rs,
	}

	return rpcServer
}

func (rpc *RpcServer) ListStrategy(context.Context, *pb.Empty) (*pb.TaskListResponse, error) {
	var ss []model.Strategy
	err := rpc.engine.Where("is_delete=0").Find(&ss)
	if err != nil {
		return nil, err
	}

	rpc.logger.Debug(ss)

	tasks := make([]*pb.Task, len(ss))
	for i, s := range ss {
		tasks[i] = &pb.Task{
			Id:       s.Id,
			Type:     pb.Type(s.Mode),
			Status:   0,
			Context:  s.Context,
			Interval: s.Interval,
		}
	}

	tp := &pb.TaskListResponse{
		Code:  true,
		Tasks: tasks,
	}
	return tp, nil
}

func (rpc *RpcServer) SendTaskResult(ctx context.Context, metrics *pb.Metrics) (*pb.TaskResultResponse, error) {
	for _, m := range metrics.Metrics {
		for _, r := range rpc.rs {
			go r.Recv(m)
		}
	}
	return &pb.TaskResultResponse{
		Code: true,
	}, nil
}

func (rpc *RpcServer) Start(ctx context.Context) {
	endpoint := fmt.Sprintf("%s:%d", rpc.config.Addr, rpc.config.RpcPort)
	rpc.logger.Info("[rpc] ", "start rpc server ", endpoint)
	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		rpc.logger.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterStrategyServiceServer(grpcServer, rpc)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			rpc.logger.Fatalf("[rpc]", "failed to serve: %v", err)
		}
	}()

	go rpc.startReceivers(ctx)

	select {
	case <-ctx.Done():
		rpc.logger.Debug("[rpc]", "graceful shut down rpc server")
		grpcServer.GracefulStop()
	}
}

func (rpc *RpcServer) startReceivers(ctx context.Context) {
	for _, r := range rpc.rs {
		for i := 0; i < RECEIVER_WORKER_COUNT; i++ {
			go r.PushLoop(ctx)
		}
	}
}

func (rpc *RpcServer) Stop() {

}
