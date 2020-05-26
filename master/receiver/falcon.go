package receiver

import (
	"context"
	"detect/master/conf"
	pb "detect/protos"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	prefix   = ""
	endpoint = "url.monitor"
)

//TODO need implement
type FalconClient interface {
	DefineGaugeMetric(string)
	EmitGaugeMetric()
}

type FalconReceiver struct {
	Client      FalconClient
	recvChan    chan *pb.Metric
	StorageConf conf.StorageConf
	Logger      *logrus.Logger
}

func NewFalconReceiver(config conf.StorageConf, Logger *logrus.Logger) *FalconReceiver {
	r := &FalconReceiver{
		recvChan:    make(chan *pb.Metric),
		StorageConf: config,
		Logger:      Logger,
	}
	//r.Client = newFalconClinet()
	return r
}

func (d *FalconReceiver) PushLoop(ctx context.Context) {
	caps := 8
	rets := make([]*pb.Metric, 0, caps)
	t := time.NewTicker(time.Second * time.Duration(10))
	for {
		select {
		case <-t.C:
			d.Logger.Debug("[receiver] time's up ! start push db")
			if len(rets) > 0 {
				d.pushRet(rets)
				rets = make([]*pb.Metric, 0, caps)
			}
		case metric := <-d.recvChan:
			d.Logger.Debug("[receiver] channel receive! start push result")
			rets = append(rets, metric)
			if len(rets) > caps {
				d.pushRet(rets)
				rets = make([]*pb.Metric, 0, caps)
			}
		case <-ctx.Done():
			d.Logger.Info("[receiver] [result channel] stop")
			return
		}
	}
}

func (d *FalconReceiver) Recv(m *pb.Metric) {
	d.Logger.Debug("[receiver]", " receive new metric")
	d.recvChan <- m
}

func (d *FalconReceiver) pushRet(metrics []*pb.Metric) {
	d.Logger.Debug("ready to push falcon ", metrics)
	for _, metric := range metrics {
		d.Client.DefineGaugeMetric(metric.Metric)
		d.Client.EmitGaugeMetric()
	}
}
