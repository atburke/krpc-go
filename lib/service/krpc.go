package service

import (
	"encoding/json"
	"fmt"

	"github.com/atburke/krpc-go/api"
	"github.com/atburke/krpc-go/lib/client"
	"github.com/atburke/krpc-go/lib/utils"
	"github.com/golang/protobuf/proto"
	"github.com/ztrue/tracerr"
)

type KRPC struct {
	client *client.KRPCClient
}

func NewKRPC(client *client.KRPCClient) *KRPC {
	return &KRPC{client: client}
}

func (s *KRPC) GetStatus() (*api.Status, error) {
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

func (s *KRPC) GetServices() (*api.Services, error) {
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

func (s *KRPC) AddStream(call *api.ProcedureCall, startStream bool) (uint64, error) {
	rawCall, err := proto.Marshal(call)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	request := &api.ProcedureCall{
		Service:   "KRPC",
		Procedure: "AddStream",
		Arguments: []*api.Argument{
			{
				Position: 0,
				Value:    rawCall,
			},
			{
				Position: 1,
				Value:    utils.EncodeBool(startStream),
			},
		},
	}
	result, err := s.client.Call(request, true)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	var stream api.Stream
	if err := proto.Unmarshal(result.Value, &stream); err != nil {
		return 0, tracerr.Wrap(err)
	}
	if err := s.client.AddStream(stream.Id); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return stream.Id, nil
}

func (s *KRPC) StartStream(id uint64) error {
	request := &api.ProcedureCall{
		Service:   "KRPC",
		Procedure: "StartStream",
		Arguments: []*api.Argument{
			{
				Position: 0,
				Value:    proto.EncodeVarint(id),
			},
		},
	}
	_, err := s.client.Call(request, false)
	return tracerr.Wrap(err)
}

func (s *KRPC) RemoveStream(id uint64) error {
	request := &api.ProcedureCall{
		Service:   "KRPC",
		Procedure: "RemoveStream",
		Arguments: []*api.Argument{
			{
				Position: 0,
				Value:    proto.EncodeVarint(id),
			},
		},
	}
	_, err := s.client.Call(request, false)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return tracerr.Wrap(s.client.RemoveStream(id))
}
