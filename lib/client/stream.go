package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/atburke/krpc-go/api"
	"github.com/atburke/krpc-go/lib/utils"
	"github.com/golang/protobuf/proto"
	"github.com/ztrue/tracerr"
)

// StreamClient is a client for kRPC streams.
type StreamClient struct {
	sync.RWMutex
	conn    net.Conn
	streams map[uint64]*streamManager
}

// NewStreamClient creates a new stream client with an existing connection.
func NewStreamClient(conn net.Conn) *StreamClient {
	return &StreamClient{
		conn:    conn,
		streams: make(map[uint64]*streamManager),
	}
}

// Close closes the stream client.
func (s *StreamClient) Close() error {
	return tracerr.Wrap(s.conn.Close())
}

// Send sends protobuf-encoded data to a stream server.
func (s *StreamClient) Send(data []byte) error {
	return tracerr.Wrap(send(s.conn, data))
}

// Receive receives protobuf-encoded data from a stream server.
func (s *StreamClient) Receive() ([]byte, error) {
	data, err := receive(s.conn)
	return data, tracerr.Wrap(err)
}

// Run starts the stream handler.
func (s *StreamClient) Run(ctx context.Context) {
	for {
		data, err := s.Receive()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stream: %v\n", err)
		}

		var streamUpdate api.StreamUpdate
		if err := proto.Unmarshal(data, &streamUpdate); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling stream result: %v\n", err)
		}
		for _, result := range streamUpdate.Results {
			s.WriteToStream(result.Id, result.Result.Value)
		}

		select {
		case <-ctx.Done():
			s.Close()
			return
		default:
		}
	}
}

// WriteToStream writes data to a particular stream.
func (s *StreamClient) WriteToStream(id uint64, b []byte) {
	s.RLock()
	sm, ok := s.streams[id]
	s.RUnlock()

	if !ok {
		s.Lock()
		// Another thread may have already created the stream manager.
		sm, ok = s.streams[id]
		if !ok {
			sm = newStreamManager()
			s.streams[id] = sm
		}
		s.Unlock()
	}

	sm.write(b)
}

// GetStream gets a byte stream for a particular stream ID.
func (s *StreamClient) GetStream(id uint64) *Stream[[]byte] {
	return s.streams[id].newStream()
}

type streamManager struct {
	channels map[int]chan []byte
	newID    func() int
	sync.RWMutex
}

func newStreamManager() *streamManager {
	return &streamManager{
		channels: make(map[int]chan []byte),
		newID:    utils.NewIDGenerator(),
	}
}

func (sm *streamManager) newStream() *Stream[[]byte] {
	sm.Lock()
	defer sm.Unlock()

	c := make(chan []byte)
	idx := sm.newID()
	sm.channels[idx] = c
	return &Stream[[]byte]{
		C:     c,
		clone: sm.newStream,
		close: func() { sm.deleteStream(idx) },
	}
}

func (sm *streamManager) deleteStream(idx int) {
	sm.Lock()
	defer sm.Unlock()

	delete(sm.channels, idx)
}

func (sm *streamManager) write(b []byte) {
	sm.RLock()
	defer sm.RUnlock()

	for _, ch := range sm.channels {
		select {
		case ch <- b:
		// Don't update channel if no one is listening.
		default:
		}
	}
}

// Stream is a struct for receiving stream data.
type Stream[T any] struct {
	C     chan T
	clone func() *Stream[T]
	close func()
}

// Clone clones the stream for another thread to listen on.
func (s *Stream[T]) Clone() *Stream[T] {
	return s.clone()
}

// Close closes the stream.
func (s *Stream[T]) Close() error {
	s.close()
	return nil
}

// MapStream converts a stream to another type.
func MapStream[S, T any](src *Stream[S], m func(S) T) *Stream[T] {
	ctx, cancel := context.WithCancel(context.Background())
	dst := &Stream[T]{
		C: make(chan T),
		clone: func() *Stream[T] {
			return MapStream(src.Clone(), m)
		},
		close: func() {
			cancel()
			src.close()
		},
	}

	go func() {
		for {
			select {
			case data := <-src.C:
				dst.C <- m(data)
			case <-ctx.Done():
				return
			}
		}
	}()

	return dst
}
