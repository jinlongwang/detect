package exec

import (
	"bytes"
	pb "detect/protos"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpExecutor struct {
	TaskId   int64
	Url      string
	Timeout  int64
	Hosts    string
	Headers  map[string]interface{}
	Cookies  map[string]interface{}
	Method   string
	Body     map[string]interface{}
	Interval int64
	Hostname string
}

func NewHttpExecutor(task *pb.Task) *HttpExecutor {
	info := make(map[string]interface{})
	err := json.Unmarshal([]byte(task.Context), &info)
	if err != nil {
		logrus.Error("[minion exe] ", "create http executor error", err)
	}

	u := info["url"].(string)
	timeout := int64(info["timeout"].(float64))
	hosts := info["hosts"].(string)
	headers := info["headers"].(map[string]interface{})
	cookies := info["cookies"].(map[string]interface{})
	body := info["body"].(map[string]interface{})
	method := info["method"].(string)
	h := &HttpExecutor{
		TaskId:   task.Id,
		Url:      u,
		Timeout:  timeout,
		Hosts:    hosts,
		Headers:  headers,
		Cookies:  cookies,
		Method:   method,
		Body:     body,
		Interval: task.Interval,
		Hostname: "",
	}
	return h
}

func (h *HttpExecutor) String() string {
	return fmt.Sprintf("url: %s, task_id: %d, interval: %d", h.Url, h.TaskId, h.Interval)
}

func (h *HttpExecutor) Execute() ([]*pb.Metric, error) {
	c, req, err := h.PreExe()
	if err != nil {
		logrus.Error("[minion exe] ", "prepare exe request error", err)
		return nil, err
	}

	begin := time.Now().UnixNano()
	res, err := c.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
	} else {
		logrus.Error("[minion exe] ", "send http error", err)
		return nil, err
	}
	end := time.Now().UnixNano()
	latency := float32(end-begin) / 1e6
	metricCode, err := h.BuildMetric(MetricHttpCode, float32(res.StatusCode), "")
	if err != nil {
		logrus.Error("[minion exe] ", "build metric ", MetricHttpCode, err)
		return nil, err
	}

	metricLatency, err := h.BuildMetric(MetricHttpLatency, latency, "")
	if err != nil {
		logrus.Error("[minion exe] ", "build metric ", MetricHttpLatency, err)
		return nil, err
	}

	return []*pb.Metric{
		metricCode,
		metricLatency,
	}, nil

}

func (h *HttpExecutor) BuildMetric(name string, value float32, t string) (*pb.Metric, error) {
	u, err := url.Parse(h.Url)
	if err != nil {
		return nil, err
	}
	tags := map[string]string{
		"hostname": h.Hostname,
		"scheme":   u.Scheme,
		"domain":   u.Host,
		"path":     u.Path,
		"query":    u.RawQuery,
	}

	m := &pb.Metric{
		StrategyId: h.TaskId,
		Metric:     name,
		Value:      value,
		Type:       t,
		Timestamp:  time.Now().UnixNano() / 1e9,
		Step:       h.Interval,
		Tags:       tags,
	}

	return m, nil
}

func (h *HttpExecutor) PushResult() {
}

func (h *HttpExecutor) GetInterval() int64 {
	if h.Interval == 0 {
		h.Interval = 30
	}
	return h.Interval
}

func (h *HttpExecutor) PreExe() (*http.Client, *http.Request, error) {
	if h.Timeout == 0 {
		h.Timeout = 30
	}

	client := &http.Client{
		Timeout: time.Duration(h.Timeout) * time.Second,
	}

	if !strings.HasPrefix(h.Url, "http") && !strings.HasPrefix(h.Url, "https") {
		return nil, nil, errors.New("protocol error")
	}

	var req *http.Request
	if h.Body != nil {
		body, _ := json.Marshal(h.Body)
		req, _ = http.NewRequest(h.Method, h.Url, bytes.NewBuffer(body))
	} else {
		req, _ = http.NewRequest(h.Method, h.Url, nil)
	}

	if h.Headers != nil {
		for name, value := range h.Headers {
			req.Header.Add(name, value.(string))
		}
	}

	if h.Hosts != "" {
		req.Host = h.Hosts
	}

	if h.Cookies != nil {
		for name, value := range h.Cookies {
			c := &http.Cookie{
				Name:  name,
				Value: value.(string),
			}
			req.AddCookie(c)
		}
	}
	return client, req, nil
}
