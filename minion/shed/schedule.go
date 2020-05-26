package shed

import "detect/minion/exec"

type Job struct {
	Offset     int64
	OffsetWait bool
	Delay      bool
	Running    bool
	Executor   exec.Executor
}
