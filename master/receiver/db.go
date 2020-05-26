package receiver

import (
	"context"
	"detect/master/conf"
	"detect/master/model"
	pb "detect/protos"
	"encoding/json"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"time"
)

type DBReceiver struct {
	recvChan    chan *pb.Metric
	StorageConf conf.StorageConf
	Logger      *logrus.Logger
	engine      *xorm.Engine
}

func (d *DBReceiver) PushLoop(ctx context.Context) {
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

func (d *DBReceiver) Recv(m *pb.Metric) {
	d.Logger.Debug("[receiver]", " receive new metric")
	d.recvChan <- m
}

func (d *DBReceiver) pushRet(metrics []*pb.Metric) {
	d.Logger.Debug(metrics)
	ms := make([]model.Metrics, 0, 0)
	for _, metric := range metrics {
		var tags string
		tagsByte, err := json.Marshal(metric.Tags)
		if err != nil {
			tags = ""
		} else {
			tags = string(tagsByte)
		}
		m := model.Metrics{
			StrategyId: metric.StrategyId,
			Metric:     fmt.Sprintf("%s.%s", prefix, metric.Metric),
			Value:      float64(metric.Value),
			Step:       metric.Step,
			MType:      metric.Type,
			Timestamp:  metric.Timestamp,
			Tags:       tags,
		}
		ms = append(ms, m)
	}
	_, err := d.engine.Insert(&ms)
	if err != nil {
		d.Logger.Error("insert metrics error ", err)
	}
}
