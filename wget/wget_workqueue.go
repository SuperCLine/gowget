package wget

import "sync"

type Work interface {
	Process()
}

type DefaultWork struct {

	worker func()
}

func NewDefaultWork(f func()) *DefaultWork {

	return &DefaultWork{
		worker:f,
	}
}

func (dw *DefaultWork) Process()  {

	dw.worker()
}


type WorkQueue struct {

	mThreadNum int
	mExit chan bool
	mTaskQueue chan Work
	mWaiter sync.WaitGroup
	mTaskWaiter sync.WaitGroup
}

func NewWorkQueue() *WorkQueue  {

	return &WorkQueue{

	}
}

func (wq *WorkQueue) Init(threadNum int, queueSize int)  {

	wq.mThreadNum = threadNum
	wq.mExit = make(chan bool, threadNum)
	wq.mWaiter = sync.WaitGroup{}
	wq.mTaskWaiter = sync.WaitGroup{}
	wq.mWaiter.Add(threadNum)
	wq.mTaskQueue = make(chan Work, queueSize)

	for i:=0; i<threadNum; i++ {

		go func() {

			defer wq.mWaiter.Done()

			for {
				select {
				case <-wq.mExit:
					return
				case task:=<-wq.mTaskQueue:
					task.Process()
					wq.mTaskWaiter.Done()
				}
			}
		}()
	}
}

func (wq *WorkQueue) Destroy()  {

	wq.WaitAllTask()

	for i:=0; i<wq.mThreadNum; i++ {
		wq.mExit <- true
	}
	wq.mWaiter.Wait()

	close(wq.mTaskQueue)
	close(wq.mExit)
}

func (wq *WorkQueue) AddTask(task Work)  {

	wq.mTaskQueue <- task
	wq.mTaskWaiter.Add(1)
}

func (wq *WorkQueue) WaitAllTask()  {

	wq.mTaskWaiter.Wait()
}