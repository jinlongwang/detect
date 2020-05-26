package http

func (h *HttpServer) registerRouter() {
	h.r.GET("/health", h.Health)
	api := h.r.Group("/api")
	{
		api.POST("strategy", h.CreateStrategy)
	}
}
