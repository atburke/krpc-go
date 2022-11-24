package krpcgo

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/atburke/krpc-go/lib/api"
	"github.com/golang/protobuf/proto"
	"github.com/ztrue/tracerr"
)

// KRPCClient is a client for a kRPC server.
type KRPCClient struct {
	mu sync.Mutex
	KRPCClientConfig
	conn net.Conn
	*StreamClient
	clientIdentifier [16]byte
}

// KRPCClientConfig is the config for a kRPC client.
type KRPCClientConfig struct {
	// Host is the kRPC server host. Defaults to "localhost".
	Host string
	// RPCPort is the kRPC server port. Defaults to "50000".
	RPCPort string
	// StreamPort is the stream server port. Defaults to "50001".
	StreamPort string
	// ClientName is the client name sent to the kRPC server. Defaults to "krpc-go".
	ClientName string
	// RPCOnly will only set up the RPC client (and not the stream client) when enabled.
	// Disabled by default.
	RPCOnly bool
}

// SetDefaults sets the config defaults.
func (cfg *KRPCClientConfig) SetDefaults() {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.RPCPort == "" {
		cfg.RPCPort = "50000"
	}
	if cfg.StreamPort == "" {
		cfg.StreamPort = "50001"
	}
	if cfg.ClientName == "" {
		cfg.ClientName = "krpc-go"
	}
}

// NewKRPCClient creates a new client.
func NewKRPCClient(cfg KRPCClientConfig) *KRPCClient {
	cfg.SetDefaults()
	return &KRPCClient{
		KRPCClientConfig: cfg,
	}
}

// Connect connects to a kRPC server.
func (c *KRPCClient) Connect(ctx context.Context) error {
	if err := c.connectRPC(); err != nil {
		return tracerr.Wrap(err)
	}
	if !c.RPCOnly {
		if err := c.connectStream(ctx); err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

// connectRPC performs the kRPC connection handshake with the RPC server.
func (c *KRPCClient) connectRPC() error {
	conn, err := net.Dial("tcp", net.JoinHostPort(c.Host, c.RPCPort))
	if err != nil {
		return tracerr.Wrap(err)
	}
	c.conn = conn

	request := api.ConnectionRequest{
		Type:       api.ConnectionRequest_RPC,
		ClientName: c.ClientName,
	}
	out, err := proto.Marshal(&request)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if err := c.Send(out); err != nil {
		return tracerr.Wrap(err)
	}
	in, err := c.Receive()
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

// connectStream creates a new stream from a kRPC client.
func (c *KRPCClient) connectStream(ctx context.Context) error {
	conn, err := net.Dial("tcp", net.JoinHostPort(c.Host, c.StreamPort))
	if err != nil {
		tracerr.Wrap(err)
	}

	request := api.ConnectionRequest{
		Type:             api.ConnectionRequest_STREAM,
		ClientIdentifier: c.clientIdentifier[:],
	}
	out, err := proto.Marshal(&request)
	if err != nil {
		tracerr.Wrap(err)
	}
	if err := send(conn, out); err != nil {
		tracerr.Wrap(err)
	}
	in, err := receive(conn)
	if err != nil {
		tracerr.Wrap(err)
	}

	var resp api.ConnectionResponse
	if err := proto.Unmarshal(in, &resp); err != nil {
		tracerr.Wrap(err)
	}
	if resp.Status != api.ConnectionResponse_OK {
		tracerr.Errorf(resp.Message)
	}

	c.StreamClient = NewStreamClient(conn)
	go c.StreamClient.Run(ctx)
	return nil
}

// Close closes the client.
func (c *KRPCClient) Close() error {
	var errors []error
	if c.StreamClient != nil {
		errors = append(errors, c.StreamClient.Close())
	}
	errors = append(errors, c.conn.Close())
	if len(errors) > 0 {
		return tracerr.Errorf("Failed to close connection(s): %v", errors)
	}
	return nil
}

// send writes length-encoded data to a writer.
func send(w io.Writer, data []byte) error {
	rawLength := proto.EncodeVarint((uint64)(len(data)))
	_, err := w.Write(rawLength)
	if err != nil {
		return tracerr.Wrap(err)
	}
	_, err = w.Write(data)
	return tracerr.Wrap(err)
}

// receive reads length-encoded data from a reader.
func receive(r io.Reader) ([]byte, error) {
	messageLength, err := readMessageLength(r)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	data := make([]byte, messageLength)
	_, err = io.ReadFull(r, data)
	return data, tracerr.Wrap(err)
}

// Send sends protobuf-encoded data to a kRPC server.
func (c *KRPCClient) Send(data []byte) error {
	return tracerr.Wrap(send(c.conn, data))
}

// Receive receives protobuf-encoded data from a kRPC server.
func (c *KRPCClient) Receive() ([]byte, error) {
	data, err := receive(c.conn)
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

// CallMultiple performs a batch of procedure calls to the rpc server.
func (c *KRPCClient) CallMultiple(calls []*api.ProcedureCall, expectResponse bool) ([]*api.ProcedureResult, error) {
	req := &api.Request{
		Calls: calls,
	}
	out, err := proto.Marshal(req)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// Lock here to prevent RPC requests from intermingling.
	c.mu.Lock()
	if err := c.Send(out); err != nil {
		c.mu.Unlock()
		return nil, tracerr.Wrap(err)
	}
	if !expectResponse {
		c.mu.Unlock()
		return nil, nil
	}
	in, err := c.Receive()
	c.mu.Unlock()

	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var resp api.Response
	if err := proto.Unmarshal(in, &resp); err != nil {
		return nil, tracerr.Wrap(err)
	}

	if resp.Error != nil {
		return nil, tracerr.Wrap(resp.Error)
	}
	return resp.Results, nil
}

// Call performs a remote procedure call.
func (c *KRPCClient) Call(call *api.ProcedureCall, expectResponse bool) (*api.ProcedureResult, error) {
	resp, err := c.CallMultiple([]*api.ProcedureCall{call}, expectResponse)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	r := resp[0]
	if r.Error != nil {
		return nil, tracerr.Wrap(r.Error)
	}
	return r, nil
}
