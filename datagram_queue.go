package quic

import (
	"sync"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

type datagramQueueEntry struct {
	datagram *wire.DatagramFrame
	sent     chan struct{}
}

type datagramQueue struct {
	rcvQueue  chan []byte
	mutex     sync.Mutex
	sendQueue []datagramQueueEntry

	closeErr error
	closed   chan struct{}

	hasData func()

	logger utils.Logger
}

func newDatagramQueue(hasData func(), logger utils.Logger) *datagramQueue {
	return &datagramQueue{
		rcvQueue: make(chan []byte, protocol.DatagramRcvQueueLen),
		hasData:  hasData,
		closed:   make(chan struct{}),
		logger:   logger,
	}
}

// AddAndWait queues a new DATAGRAM frame for sending.
// It blocks until the frame has been dequeued.
func (h *datagramQueue) AddAndWait(f *wire.DatagramFrame) error {
	c := make(chan struct{})
	h.mutex.Lock()
	h.sendQueue = append(h.sendQueue, datagramQueueEntry{
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

// Get dequeues a DATAGRAM frame for sending.
func (h *datagramQueue) Get() *wire.DatagramFrame {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if len(h.sendQueue) == 0 {
		return nil
	}
	e := h.sendQueue[0]
	h.sendQueue = h.sendQueue[1:]
	close(e.sent)
	return e.datagram
}

// HandleDatagramFrame handles a received DATAGRAM frame.
func (h *datagramQueue) HandleDatagramFrame(f *wire.DatagramFrame) {
	data := make([]byte, len(f.Data))
	copy(data, f.Data)
	select {
	case h.rcvQueue <- data:
	default:
		h.logger.Debugf("Discarding DATAGRAM frame (%d bytes payload)", len(f.Data))
	}
}

// Receive gets a received DATAGRAM frame.
func (h *datagramQueue) Receive() ([]byte, error) {
	select {
	case data := <-h.rcvQueue:
		return data, nil
	case <-h.closed:
		return nil, h.closeErr
	}
}

func (h *datagramQueue) CloseWithError(e error) {
	h.closeErr = e
	close(h.closed)
}
