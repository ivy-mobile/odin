package queue

// Queue 基于channel的简单队列
type Queue struct {
	ch   chan func()
	done chan struct{} // 关闭信号
}

func New(size int) *Queue {
	return &Queue{
		ch:   make(chan func(), size),
		done: make(chan struct{}),
	}
}

// Enqueue 入队
func (q *Queue) Enqueue(f func()) {
	q.ch <- f
}

// Dequeue 出队
func (q *Queue) Dequeue() (func(), bool) {
	v, ok := <-q.ch
	return v, ok
}

// Chan 获取队列的channel
// 当select有其它case时,使用Dequeue() 会造成阻塞,推荐如下方式:
//
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case f := <-q.Chan():
//			f()
//		}
//	}
func (q *Queue) Chan() chan func() {
	return q.ch
}

// Done 获取队列的结束信号
func (q *Queue) Done() chan struct{} {
	return q.done
}

// Close 关闭队列
func (q *Queue) Close() {
	close(q.done)
	close(q.ch)
}
