package internal

import (
	"encoding/json"
	"fmt"

	"github.com/atburke/krpc-go/api"
	"github.com/atburke/krpc-go/lib/client"
	"github.com/golang/protobuf/proto"
	"github.com/ztrue/tracerr"
)

// BasicKRPC is a partial implementation of the KRPC service. This should only
// be used to fetch the rest of the services.
type BasicKRPC struct {
	client *client.KRPCClient
}

func NewBasicKRPC(client *client.KRPCClient) *BasicKRPC {
	return &BasicKRPC{client: client}
}

func (s *BasicKRPC) GetStatus() (*api.Status, error) {
	request := &api.ProcedureCall{
		Service:   "KRPC",
		Procedure: "GetStatus",
	}
	result, err := s.client.Call(request, true)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var status api.Status
	if err := proto.Unmarshal(result.Value, &status); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &status, nil
}

func (s *BasicKRPC) GetServices() (*api.Services, error) {
	fmt.Println("GetServices")
	request := &api.ProcedureCall{
		Service:   "KRPC",
		Procedure: "GetServices",
	}
	o, _ := json.Marshal(request)
	fmt.Println(string(o))
	result, err := s.client.Call(request, true)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var services api.Services
	if err := proto.Unmarshal(result.Value, &services); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &services, nil
}
