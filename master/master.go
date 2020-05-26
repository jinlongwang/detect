package master

import (
	"context"
	"detect/master/conf"
	"detect/master/model"
	"detect/master/service/http"
	"detect/master/service/rpc"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"sync"
)

type Master struct {
	Logger     *logrus.Logger
	Config     *conf.MasterConfig
	context    context.Context
	cancelFunc context.CancelFunc
	wg         *sync.WaitGroup
	httpServer *http.HttpServer
	rpcServer  *rpc.RpcServer
	engine     *xorm.Engine
}

func NewMaster(logger *logrus.Logger, config *conf.MasterConfig) *Master {
	ctx, cancelFunc := context.WithCancel(context.Background())
	engine, err := model.InitSqlEngine(config.Mysql, "mysql")
	if err != nil {
		logger.Fatal("connect mysql error", err)
	}
	h := http.NewHttpService(config, logger, engine)
	r := rpc.NewRpcServer(engine, config, logger)
	master := &Master{
		context:    ctx,
		cancelFunc: cancelFunc,
		Logger:     logger,
		Config:     config,
		wg:         &sync.WaitGroup{},
		httpServer: h,
		rpcServer:  r,
		engine:     engine,
	}
	return master
}

func (m *Master) Start() {
	m.wg.Add(1)
	go func(ctx context.Context) {
		defer m.wg.Done()
		m.httpServer.Start(m.context)
	}(m.context)

	m.wg.Add(1)
	go func(ctx context.Context) {
		defer m.wg.Done()
		m.rpcServer.Start(m.context)
	}(m.context)

}

func (m *Master) Stop() {
	m.cancelFunc()
	m.wg.Wait()
}
