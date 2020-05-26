package http

import (
	"detect/master/model"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *HttpServer) Health(c *gin.Context) {
	results, err := h.engine.Query("select * from os_version")
	if err != nil {
		h.logger.Debug("123213")
		h.logger.Debug(err)
	}
	h.logger.Debug(results)
	c.JSON(200, gin.H{
		"message": "ok",
	})
}

func (h *HttpServer) CreateStrategy(c *gin.Context) {
	var s StrategyJson
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	context, err := json.Marshal(s.Context)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	strategy := &model.Strategy{
		Name:     s.Name,
		Note:     s.Note,
		Mode:     s.Mode,
		IsDelete: s.IsDelete,
		Context:  string(context),
	}

	_, err = h.engine.Insert(strategy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug(s)

	c.JSON(200, gin.H{
		"message": "ok",
		"s_id":    strategy.Id,
	})
}
