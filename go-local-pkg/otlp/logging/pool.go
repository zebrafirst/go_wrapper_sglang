package logging

import (
	"context"
	"git.iflytek.com/AIaaS/otlp-self/v3/monitor"
	"git.iflytek.com/AIaaS/otlp-self/v3/utils"
	"git.iflytek.com/AIaaS/otlp-self/v3/zaplog"
)

type job interface {
	Do()
}

type worker struct {
	id   int
	name string
}

type TaskFunc func()

type Task struct {
	f TaskFunc
}

func (t *Task) Do() {
	t.f()
}

func newWorker(id int, name string) *worker {
	return &worker{
		id:   id,
		name: name,
	}
}

func (w *worker) run(ctx context.Context, jobQueue chan job) {
	go func() {
		defer utils.Catch("pool run")
		for {
			select {
			case <-ctx.Done():
				return
			case j := <-jobQueue:
				j.Do()
			}
		}
	}()
}

type WorkerPool struct {
	ctx     context.Context
	cancel  context.CancelFunc
	name    string
	workers []*worker
	queue   chan job

	block   bool
	maxSize int
}

func newWorkerPool(name string, queueSize, workerNum int, block bool) *WorkerPool {
	if workerNum <= 0 {
		workerNum = 10
	}
	if queueSize <= 0 {
		queueSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		ctx:     ctx,
		name:    name,
		cancel:  cancel,
		block:   block,
		maxSize: queueSize,
		workers: func() []*worker {
			workers := make([]*worker, 0, workerNum)
			for i := 0; i < workerNum; i++ {
				workers = append(workers, newWorker(i, "default"))
			}
			return workers
		}(),
		queue: make(chan job, queueSize),
	}
}

func (wp *WorkerPool) start() *WorkerPool {
	for _, w := range wp.workers {
		w.run(wp.ctx, wp.queue)
	}
	return wp
}

func (wp *WorkerPool) AppendJob(j job) {
	if !wp.block && (len(wp.queue) == wp.maxSize) {
		zaplog.SDKLogger.Errorf("%v discard due to not block and queue equals maxsize: %v", wp.name, wp.maxSize)
		if monitor.SDKMetricsEnable() {
			monitor.OtlpSdkMetrics.RecordCounter(wp.name+"_exceed", 1)
		}
		return
	}
	wp.queue <- j
}

func (wp *WorkerPool) stop() {
	wp.cancel()
}
