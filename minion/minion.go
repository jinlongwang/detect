package minion

import (
	"context"
	"detect/minion/conf"
	"detect/minion/shed"
	pb "detect/protos"
	"github.com/benbjohnson/clock"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"sync"
	"time"
)

const (
	MAX_GORUTINE_COUNT       = 10
	MAX_RESULT_CHANNEL_COUNT = 1000
)

type Minion struct {
	Config         *conf.AgentConfig
	Logger         *logrus.Logger
	executeChan    chan *shed.Job
	resultChan     chan *pb.Metric
	ExecutorShed   *shed.FetchSchedule
	StrategyClient pb.StrategyServiceClient
	wg             *sync.WaitGroup
	context        context.Context
	cancelFunc     context.CancelFunc
	conn           *grpc.ClientConn
	ticker         *shed.Ticker
}

func NewMinion(logger *logrus.Logger, config *conf.AgentConfig) *Minion {
	ctx, cancelFunc := context.WithCancel(context.Background())
	minion := &Minion{
		Config:      config,
		Logger:      logger,
		executeChan: make(chan *shed.Job),
		resultChan:  make(chan *pb.Metric, MAX_RESULT_CHANNEL_COUNT),
		context:     ctx,
		cancelFunc:  cancelFunc,
		wg:          &sync.WaitGroup{},
		ticker:      shed.NewTicker(time.Now(), time.Second*0, clock.New()),
	}
	minion.initRpcClient()
	exeShed := shed.NewFetchSchedule(config.Interval, minion.StrategyClient, logger)
	minion.ExecutorShed = exeShed
	return minion
}

func (m *Minion) initRpcClient() {
	conn, err := grpc.Dial(m.Config.MasterAddr, grpc.WithInsecure())
	if err != nil {
		m.Logger.Fatal("[rpc_client] connect to master error", err)
	}
	m.conn = conn
	strategyClient := pb.NewStrategyServiceClient(conn)
	m.StrategyClient = strategyClient
}

func (m *Minion) Start() {
	defer func() {
		if err := recover(); err != nil {
			m.Logger.Error("minion Panic: stopping minion", "error", err, "stack")
		}
	}()

	go func() {
		tickIndex := 0
		for {
			select {
			case <-m.context.Done():
				m.Logger.Debug("[minion] [shed]", "stop shed loop")
				return
			case tick := <-m.ticker.C:
				// TEMP SOLUTION update rules ever tenth tick
				if tickIndex%10 == 0 {
					m.Logger.Debug("[minion] [shed]", "start update strategy")
					m.ExecutorShed.Run(m.context)
				}

				m.ExecutorShed.Tick(tick, m.executeChan)
				tickIndex++
			}
		}
	}()

	go m.StartExeLoop()
	go m.WaitResult()
}

func (m *Minion) StartExeLoop() {
	m.wg.Add(MAX_GORUTINE_COUNT)
	for i := 0; i < MAX_GORUTINE_COUNT; i++ {
		go func(ctx context.Context, i int) {
			m.Logger.Debug("[minion] [exeloop] ", "start execute worker id: ", i)
			defer m.wg.Done()
			for {
				select {
				case job := <-m.executeChan:
					job.Running = true
					metrics, err := job.Executor.Execute()
					job.Running = false
					if err == nil {
						for _, metric := range metrics {
							m.Logger.Debug("push to result channel ", metric.Metric)
							m.resultChan <- metric
						}
					}
				case <-ctx.Done():
					m.Logger.Debug("[minion] [exeloop] ", "stop execute worker id: ", i)
					return
				}
			}
		}(m.context, i)
	}
}

func (m *Minion) WaitResult() {
	caps := 8
	rets := make([]*pb.Metric, 0, caps)
	t := time.NewTicker(time.Second * time.Duration(10))
	for {
		select {
		case <-t.C:
			m.Logger.Debug("[minion] [result channel] time up! start push result")
			if len(rets) > 0 {
				m.uploadMetrics(rets)
				rets = make([]*pb.Metric, 0, caps)
			}
		case metric := <-m.resultChan:
			m.Logger.Debug("[minion] [result channel] channel receive! start push result")
			rets = append(rets, metric)
			if len(rets) > caps {
				m.uploadMetrics(rets)
				rets = make([]*pb.Metric, 0, caps)
			}
		case <-m.context.Done():
			m.Logger.Info("[minion] [result channel] stop")
			return
		}
	}
}

func (m *Minion) uploadMetrics(metrics []*pb.Metric) {
	for _, metric := range metrics {
		m.Logger.Debug(metric.Metric, " ", metric.Value, " ", metric.Tags)
	}
	req := &pb.Metrics{
		Metrics: metrics,
	}
	res, err := m.StrategyClient.SendTaskResult(m.context, req)
	if err != nil {
		m.Logger.Error("[minion] [upload] upload metric error", err)
		return
	}
	if !res.Code {
		m.Logger.Error("[minion] [upload] upload metric error", err)
		return
	}
}

func (m *Minion) Stop() {
	m.Logger.Info("[minion] ", "minion stopping")
	m.cancelFunc()
	m.wg.Wait()
	m.conn.Close()
	m.Logger.Info("[minion] ", "minion already stopped")
}
