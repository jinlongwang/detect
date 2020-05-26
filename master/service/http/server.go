package http

import (
	"context"
	"detect/master/conf"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"net/http"
)

type HttpServer struct {
	r      *gin.Engine
	config *conf.MasterConfig
	logger *logrus.Logger
	engine *xorm.Engine
}

func (h *HttpServer) addMiddleware(handlerFunc gin.HandlerFunc) {
	h.r.Use(handlerFunc)
}

func NewHttpService(conf *conf.MasterConfig, logger *logrus.Logger, engine *xorm.Engine) *HttpServer {
	r := gin.New()
	httpServer := &HttpServer{
		r:      r,
		config: conf,
		logger: logger,
		engine: engine,
	}
	httpServer.addMiddleware(gin.Logger())
	httpServer.addMiddleware(gin.Recovery())
	httpServer.registerRouter()
	return httpServer
}

func (h *HttpServer) Start(ctx context.Context) {
	endpoint := fmt.Sprintf("%s:%d", h.config.Addr, h.config.HttpPort)
	h.logger.Info("http", "http server start in ", endpoint)

	srv := &http.Server{
		Addr:    endpoint,
		Handler: h.r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Fatalf("listen: %s\n", err)
		}
	}()

	select {
	case <-ctx.Done():
		h.logger.Debug("[http]", "graceful shut down gin server")
		if err := srv.Shutdown(ctx); err != nil {
			h.logger.Fatal("Server forced to shutdown:", err)
		}
	}
}

func (h *HttpServer) Stop() {

}
