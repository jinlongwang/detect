package receiver

import (
	"context"
	"detect/master/conf"
	pb "detect/protos"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
)

type Receiver interface {
	PushLoop(context.Context)
	Recv(*pb.Metric)
	pushRet([]*pb.Metric)
}

func NewReceivers(config *conf.MasterConfig, Logger *logrus.Logger, engine *xorm.Engine) []Receiver {
	receivers := make([]Receiver, 0, 0)
	if config.Storage.Db.Enable {
		r := &DBReceiver{
			recvChan:    make(chan *pb.Metric),
			StorageConf: config.Storage.Db,
			Logger:      Logger,
			engine:      engine,
		}
		receivers = append(receivers, r)
	}

	if config.Storage.Metrics.Enable {
		r := &MetricReceiver{}
		receivers = append(receivers, r)
	}

	if config.Storage.Falcon.Enable {
		r := NewFalconReceiver(config.Storage.Falcon, Logger)
		receivers = append(receivers, r)
	}
	return receivers
}
