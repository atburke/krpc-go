package gen

const testProcedure = `
package gentest

import (
	krpcgo "github.com/atburke/krpc-go"
	api "github.com/atburke/krpc-go/lib/api"
	encode "github.com/atburke/krpc-go/lib/encode"
	krpc "github.com/atburke/krpc-go/krpc"
	tracerr "github.com/ztrue/tracerr"
)

// MyProcedure - test procedure generation.
//
// Allowed game scenes: FLIGHT.
func (s *MyService) MyProcedure(param1 uint64, param2 string) (bool, error) {
	var err error
	var argBytes []byte
	var vv bool
	request := &api.ProcedureCall{
		Procedure: "MyProcedure",
		Service: "MyService",
	}
	argBytes, err = encode.Marshal(param1)
	if err != nil {
		return vv, tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x0),
		Value: argBytes,
	})
	argBytes, err = encode.Marshal(param2)
	if err != nil {
		return vv, tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x1),
		Value: argBytes,
	})
	result, err := s.Client.Call(request)
	if err != nil {
		return vv, tracerr.Wrap(err)
	}
	err = encode.Unmarshal(result.Value, &vv)
	if err != nil {
		return vv, tracerr.Wrap(err)
	}
	return vv, nil
}

// StreamMyProcedure - test procedure generation.
//
// Allowed game scenes: FLIGHT.
func (s *MyService) StreamMyProcedure(param1 uint64, param2 string) (*krpcgo.Stream[bool], error) {
	var err error
	var argBytes []byte
	request := &api.ProcedureCall{
		Procedure: "MyProcedure",
		Service: "MyService",
	}
	argBytes, err = encode.Marshal(param1)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x0),
		Value: argBytes,
	})
	argBytes, err = encode.Marshal(param2)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x1),
		Value: argBytes,
	})
	krpc := krpc.NewKRPC(s.Client)
	st, err := krpc.AddStream(request, true)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	rawStream := s.Client.GetStream(st.Id)
	stream := krpcgo.MapStream(rawStream, func(b []byte)bool {
		var value bool
		encode.Unmarshal(b, &value)
		return value
	})
	stream.AddCloser(func() error {
		return tracerr.Wrap(krpc.RemoveStream(st.Id))
	})
	return stream, nil
}
`

const testClassSetter = `
package gentest

import (
	api "github.com/atburke/krpc-go/lib/api"
	encode "github.com/atburke/krpc-go/lib/encode"
	tracerr "github.com/ztrue/tracerr"
)

// SetMyProperty - test class setter generation.
//
// Allowed game scenes: any.
func (s *MyClass) SetMyProperty(param1 api.Tuple2[string, uint64]) error {
	var err error
	var argBytes []byte
	request := &api.ProcedureCall{
		Procedure: "MyClass_set_MyProperty",
		Service: "MyService",
	}
	argBytes, err = encode.Marshal(s)
	if err != nil {
		return tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x0),
		Value: argBytes,
	})
	argBytes, err = encode.Marshal(param1)
	if err != nil {
		return tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x1),
		Value: argBytes,
	})
	_, err = s.Client.Call(request)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}
`
