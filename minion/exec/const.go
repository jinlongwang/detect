package exec

const (
	GET  string = "GET"
	POST string = "POST"
	HEAD string = "HEAD"

	MetricHttpCode    = "http.status"
	MetricHttpLatency = "http.latency"

	MetricTcpAlive     = "tcp.alive"
	MetricTcpConnTime  = "tcp.conn.time"
	MetricTcpWriteTime = "tcp.write.time"
	MetricTcpReadTime  = "tcp.read.time"
)
