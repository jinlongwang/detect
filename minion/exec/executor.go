package exec

import (
	pb "detect/protos"
)

type Executor interface {
	Execute() ([]*pb.Metric, error)
	BuildMetric(string, float32, string) (*pb.Metric, error)
	PushResult()
	GetInterval() int64
	String() string
}
