package client

import (
	"io"
	"net"

	"github.com/atburke/krpc-go/api"
	"github.com/golang/protobuf/proto"
	"github.com/ztrue/tracerr"
)

// KRPCClient is a client for a kRPC server.
type KRPCClient struct {
	host             string
	conn             net.Conn
	clientIdentifier [16]byte
}

// Connect connects to a kRPC server.
func Connect(clientName, host, port string) (*KRPCClient, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	c := &KRPCClient{host: host, conn: conn}
	if err := c.connectRPC(clientName); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return c, nil
}

// connectRPC performs the kRPC connection handshake with the RPC server.
func (c *KRPCClient) connectRPC(clientName string) error {
	request := api.ConnectionRequest{
		Type:       api.ConnectionRequest_RPC,
		ClientName: clientName,
	}
	out, err := proto.Marshal(&request)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if err := Send(c.conn, out); err != nil {
		return tracerr.Wrap(err)
	}
	in, err := Receive(c.conn)
	if err != nil {
		return tracerr.Wrap(err)
	}
	var resp api.ConnectionResponse
	if err := proto.Unmarshal(in, &resp); err != nil {
		return tracerr.Wrap(err)
	}
	if resp.Status != api.ConnectionResponse_OK {
		return tracerr.Errorf(resp.Message)
	}
	copy(c.clientIdentifier[:], resp.ClientIdentifier)
	return nil
}

// Close closes the client.
func (c *KRPCClient) Close() error {
	return tracerr.Wrap(c.conn.Close())
}

// Send sends protobuf-encoded data to a kRPC server.
func Send(w io.Writer, data []byte) error {
	rawLength := proto.EncodeVarint((uint64)(len(data)))
	_, err := w.Write(rawLength)
	if err != nil {
		return tracerr.Wrap(err)
	}
	_, err = w.Write(data)
	return tracerr.Wrap(err)
}

// Receive receives protobuf-encoded data from a kRPC server.
func Receive(r io.Reader) ([]byte, error) {
	messageLength, err := readMessageLength(r)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	data := make([]byte, messageLength)
	_, err = io.ReadFull(r, data)
	return data, tracerr.Wrap(err)
}

// readMessageLength attempts to read the varint-encoded length of
// a message
func readMessageLength(r io.Reader) (uint64, error) {
	var rawLength []byte
	for len(rawLength) < 16 {
		b := make([]byte, 1)
		_, err := r.Read(b)
		if err != nil {
			return 0, tracerr.Wrap(err)
		}
		rawLength = append(rawLength, b...)
		length, size := proto.DecodeVarint(rawLength)
		if size > 0 {
			return length, nil
		}
	}
	return 0, tracerr.Errorf("Message does not appear to start with length: %v", rawLength)
}

// StreamClient is a client for kRPC streams.
type StreamClient struct {
	conn net.Conn
	ch   chan []byte
}

// NewStream creates a new stream from a kRPC client.
func (c *KRPCClient) NewStream(port string) (*StreamClient, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(c.host, port))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	request := api.ConnectionRequest{
		Type:             api.ConnectionRequest_STREAM,
		ClientIdentifier: c.clientIdentifier[:],
	}
	out, err := proto.Marshal(&request)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if err := Send(conn, out); err != nil {
		return nil, tracerr.Wrap(err)
	}
	in, err := Receive(conn)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var resp api.ConnectionResponse
	if err := proto.Unmarshal(in, &resp); err != nil {
		return nil, tracerr.Wrap(err)
	}
	if resp.Status != api.ConnectionResponse_OK {
		return nil, tracerr.Errorf(resp.Message)
	}
	return &StreamClient{
		conn: conn,
		ch:   make(chan []byte),
	}, nil
}
