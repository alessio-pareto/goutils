package goutils

import (
	"fmt"
	"sync"
)

// Broadcaster provides an easy way to spread a message to
// to an unlimited number of listeners, with concurrency in mind.
// The message is not only ment as a string, but can be whatever
// type of data you want
type Broadcaster[T any] struct {
	incr int
	listeners map[int]*BroadcastListener[T]
	mutex sync.Mutex
}

// BroadcastListener is the product of the registration to a
// Broadcaster and listens for the incoming messages
type BroadcastListener[T any] struct {
	id int
	bc *Broadcaster[T]
	msg chan T
	resp chan struct{}
	state int // 0 = wating for message | 1 = message received, must report
}

// Creates a new Broadcaster
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

// Sends a message to the current registered listeners. The process is done
// in a different GoRouting, but each process waits first the spread of the
// previous message
func (bc *Broadcaster[T]) Send(msg T) {
	go bc.send(msg)
}

// Sends a message and waits all the listeners to report their usage
func (bc *Broadcaster[T]) SendAndWait(msg T) {
	bc.send(msg)
}

// Creates a new Listener
func (bc *Broadcaster[T]) Subscribe() *BroadcastListener[T] {
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

// Waits for the next message. The listener should the notify the
// Broadcaster when it has finished using the message
func (l *BroadcastListener[T]) Listen() T {
	if l.state != 0 {
		panic(fmt.Errorf("listen: previous message not handled correcly, probably missing report"))
	}
	return <- l.msg
}

// Communicates to the Broadcaster that the message has been used
func (l *BroadcastListener[T]) Report() {
	if l.state != 1 {
		panic(fmt.Errorf("report: no message waiting to be reported"))
	}
	l.resp <- struct{}{}
}

// Waits for the message and tells the Broadcaster to continue instantly
func (l *BroadcastListener[T]) Get() T {
	if l.state != 0 {
		panic(fmt.Errorf("listen: previous message not handled correcly, probably missing report"))
	}

	res := <- l.msg
	l.resp <- struct{}{}

	return res
}

// Removes the listener from the Broadcaster, making it unusable
func (l *BroadcastListener[T]) Unsubscribe() {
	if l.state != 0 {
		panic(fmt.Errorf("unsubscribe: previous message not handled correcly, cannot leave"))
	}
	close(l.msg)
	close(l.resp)

	delete(l.bc.listeners, l.id)
}
