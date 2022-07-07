package goutils

import (
	"sync"
)

type Broadcaster[T any] struct {
	incr int
	listeners map[int]*BroadcastListener[T]
	mutex sync.Mutex
}

type BroadcastListener[T any] struct {
	id int
	bc *Broadcaster[T]
	msg chan T
	resp chan struct{}
}

func NewBroadcaster[T any]() *Broadcaster[T] {
	bc := new(Broadcaster[T])
	bc.listeners = make(map[int]*BroadcastListener[T])

	return bc
}

func (bc *Broadcaster[T]) send(msg T) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	wg := new(sync.WaitGroup)
	for _, l := range bc.listeners {
		wg.Add(1)

		go func(l *BroadcastListener[T]) {
			l.msg <- msg
			<- l.resp

			wg.Done()
		}(l)
	}
	
	wg.Wait()
}

func (bc *Broadcaster[T]) Send(msg T) {
	go bc.send(msg)
}

func (bc *Broadcaster[T]) SendAndWait(msg T) {
	bc.send(msg)
}

func (bc *Broadcaster[T]) Register() *BroadcastListener[T] {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	l := &BroadcastListener[T] {
		id: bc.incr,
		bc: bc,
		msg: make(chan T, 10),
		resp: make(chan struct{}, 10),
	}

	bc.listeners[bc.incr] = l
	bc.incr ++
	
	return l
}

func (l *BroadcastListener[T]) Listen() T {
	res := <- l.msg
	l.resp <- struct{}{}

	return res
}

func (l *BroadcastListener[T]) Unregister() {
	close(l.msg)
	close(l.resp)

	delete(l.bc.listeners, l.id)
}
