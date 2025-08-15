package internal

import (
	"github.com/gammazero/workerpool"
)

type Pool struct {
	enableWait bool
	workPool   *workerpool.WorkerPool
}

func NewWorkPool(maxWorkers int, enableWait bool) *Pool {
	workerPool := workerpool.New(maxWorkers)
	return &Pool{
		workPool:   workerPool,
		enableWait: enableWait,
	}
}

func (p *Pool) Submit(task func()) {
	if task != nil && p.workPool != nil {
		//p.taskQueue <- task
		p.workPool.Submit(task)
	}
}

func (p *Pool) Stop() {
	if p.enableWait {
		p.workPool.StopWait()
	} else {
		p.workPool.Stop()
	}
}
