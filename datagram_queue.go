package quic

import (
	"sync"

	"github.com/lucas-clemente/quic-go/internal/wire"
)

type datagramQueueEntry struct {
	datagram *wire.DatagramFrame
	sent     chan struct{}
}

type datagramQueue struct {
	mutex sync.Mutex
	queue []datagramQueueEntry

	closeErr error
	closed   chan struct{}

	hasData func()
}

func newDatagramQueue(hasData func()) *datagramQueue {
	return &datagramQueue{
		hasData: hasData,
		closed:  make(chan struct{}),
	}
}

// AddAndWait queues a new DATAGRAM frame.
// It blocks until the frame has been dequeued.
func (h *datagramQueue) AddAndWait(f *wire.DatagramFrame) error {
	c := make(chan struct{})
	h.mutex.Lock()
	h.queue = append(h.queue, datagramQueueEntry{
		datagram: f,
		sent:     c,
	})
	h.mutex.Unlock()

	h.hasData()
	select {
	case <-c:
		return nil
	case <-h.closed:
		return h.closeErr
	}
}

func (h *datagramQueue) Get() *wire.DatagramFrame {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if len(h.queue) == 0 {
		return nil
	}
	e := h.queue[0]
	h.queue = h.queue[1:]
	close(e.sent)
	return e.datagram
}

func (h *datagramQueue) CloseWithError(e error) {
	h.closeErr = e
	close(h.closed)
}
