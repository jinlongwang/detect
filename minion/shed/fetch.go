package shed

import (
	"context"
	"detect/minion/exec"
	pb "detect/protos"
	"github.com/sirupsen/logrus"
	"math"
	"sync"
	"time"
)

type ExecutorCache struct {
	sync.Mutex
	jobs map[int64]*Job
	f    *FetchSchedule
}

func (e *ExecutorCache) push(index int, length int, taskId int64, exe exec.Executor) {
	defer e.Unlock()
	e.Lock()
	e.f.logger.Debug("update task")

	var job *Job
	if e.jobs[taskId] != nil {
		job = e.jobs[taskId]
	} else {
		job = &Job{
			Running: false,
		}
	}
	job.Executor = exe
	offset := ((exe.GetInterval() * 1000) / int64(length)) * int64(index) //为了打撒开始时间
	job.Offset = int64(math.Floor(float64(offset) / 1000))
	if job.Offset == 0 { //zero offset causes division with 0 panics.
		job.Offset = 1
	}
	e.jobs[taskId] = job
}

func NewExecutorCache(f *FetchSchedule) *ExecutorCache {
	return &ExecutorCache{
		Mutex: sync.Mutex{},
		jobs:  make(map[int64]*Job),
		f:     f,
	}
}

type FetchSchedule struct {
	Interval       time.Duration
	StrategyClient pb.StrategyServiceClient
	exeCache       *ExecutorCache
	logger         logrus.Logger
}

func NewFetchSchedule(interval time.Duration, s pb.StrategyServiceClient, logger *logrus.Logger) *FetchSchedule {
	f := &FetchSchedule{
		logger:         *logger,
		Interval:       interval,
		StrategyClient: s,
	}
	f.exeCache = NewExecutorCache(f)
	return f
}

func (f *FetchSchedule) Tick(tickTime time.Time, execQueue chan *Job) {
	now := tickTime.Unix()

	for _, job := range f.exeCache.jobs {
		if job.Running {
			continue
		}

		if job.OffsetWait && now%job.Offset == 0 {
			job.OffsetWait = false
			f.enqueue(job, execQueue)
			continue
		}

		if now%job.Executor.GetInterval() == 0 {
			if job.Offset > 0 {
				job.OffsetWait = true
			} else {
				f.enqueue(job, execQueue)
			}
		}
	}
}

func (f *FetchSchedule) Run(ctx context.Context) {
	taskListResponse, err := f.StrategyClient.ListStrategy(ctx, &pb.Empty{})
	if err != nil {
		f.logger.Error("[minion] [rpc client] ", "call list strategy error ", err)
		return
	}
	f.logger.Debug("[minion] [rpc client] ", taskListResponse.Code)

	if !taskListResponse.Code {
		f.logger.Error("[minion] [rpc client] ", "call list strategy empty")
		return
	}

	for i, task := range taskListResponse.Tasks {
		switch task.Type {
		case pb.Type_URL:
			h := exec.NewHttpExecutor(task)
			f.exeCache.push(i, len(taskListResponse.Tasks), h.TaskId, h)
		case pb.Type_TCP:
			h := exec.NewTcpExecutor(task)
			f.exeCache.push(i, len(taskListResponse.Tasks), h.TaskId, h)
		}
	}
}

func (f *FetchSchedule) enqueue(job *Job, execQueue chan *Job) {
	f.logger.Debug("Scheduler: Putting job on to exec queue", "name", job.Executor.String(), " ", time.Now().Unix())
	execQueue <- job
}
