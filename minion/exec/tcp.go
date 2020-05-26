package exec

import (
	pb "detect/protos"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"time"
)

type TcpExecutor struct {
	TaskId   int64
	Addr     string
	Port     int64
	Timeout  int64
	Payload  string
	Interval int64
}

func NewTcpExecutor(task *pb.Task) *TcpExecutor {
	info := make(map[string]interface{})
	err := json.Unmarshal([]byte(task.Context), &info)
	if err != nil {
		logrus.Error("[minion exe] ", "create http executor error", err)
	}
	addr := info["addr"].(string)
	timeout := int64(info["timeout"].(float64))
	port := int64(info["port"].(float64))
	payload := info["payload"].(string)
	t := &TcpExecutor{
		TaskId:   task.Id,
		Addr:     addr,
		Port:     port,
		Timeout:  timeout,
		Payload:  payload,
		Interval: task.Interval,
	}
	return t
}

func (tcp *TcpExecutor) String() string {
	return fmt.Sprintf("addr: %s:%d, task_id: %d, interval: %d", tcp.Addr, tcp.Port, tcp.TaskId, tcp.Interval)
}

func (tcp *TcpExecutor) Execute() ([]*pb.Metric, error) {
	ms := make([]*pb.Metric, 0)
	if tcp.Timeout == 0 {
		tcp.Timeout = 3
	}
	timeout := time.Duration(tcp.Timeout) * time.Second
	address := fmt.Sprintf("%s:%d", tcp.Addr, tcp.Port)
	// conn time
	begin := time.Now().UnixNano()
	conn, err := net.DialTimeout("tcp4", address, timeout)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		m, _ := tcp.BuildMetric(MetricTcpAlive, 0, "Gauge")
		return []*pb.Metric{
			m,
		}, nil
	} else {
		m, err := tcp.BuildMetric(MetricTcpAlive, 1, "Gauge")
		if err == nil {
			ms = append(ms, m)
		}
	}

	connTime := float32((time.Now().UnixNano() - begin)) / 1e6
	m, _ := tcp.BuildMetric(MetricTcpConnTime, connTime, "Gauge")
	ms = append(ms, m)

	if tcp.Payload == "" {
		return ms, nil
	}

	//set read write timeout time
	conn.SetDeadline(time.Now().Add(time.Second * 3))
	tcpConn, _ := conn.(*net.TCPConn)
	tcpConn.SetNoDelay(true)

	// send ping msg
	wrbegin := time.Now().UnixNano()
	conn.Write([]byte(tcp.Payload))
	endw := time.Now().UnixNano()

	b := make([]byte, 100)
	io.ReadFull(conn, b)
	endr := time.Now().UnixNano()

	// millisecond
	writeTime := float32((endw - wrbegin)) / 1e6
	readTime := float32((endr - endw)) / 1e6

	mw, _ := tcp.BuildMetric(MetricTcpWriteTime, writeTime, "Gauge")
	ms = append(ms, mw)

	mr, _ := tcp.BuildMetric(MetricTcpReadTime, readTime, "Gauge")
	ms = append(ms, mr)
	conn.Close()
	return ms, nil
}

func (tcp *TcpExecutor) BuildMetric(name string, value float32, t string) (*pb.Metric, error) {
	tags := map[string]string{
		"address": fmt.Sprintf("%s:%d", tcp.Addr, tcp.Port),
		"s_id":    strconv.Itoa(int(tcp.TaskId)),
	}
	m := &pb.Metric{
		StrategyId: tcp.TaskId,
		Metric:     name,
		Value:      value,
		Type:       t,
		Timestamp:  time.Now().UnixNano() / 1e9,
		Step:       tcp.Interval,
		Tags:       tags,
	}
	return m, nil
}

func (tcp *TcpExecutor) GetInterval() int64 {
	if tcp.Interval == 0 {
		tcp.Interval = 30
	}
	return tcp.Interval
}

func (tcp *TcpExecutor) PushResult() {
}
