package dockingcamera

import (
	krpcgo "github.com/atburke/krpc-go"
	krpc "github.com/atburke/krpc-go/krpc"
	api "github.com/atburke/krpc-go/lib/api"
	encode "github.com/atburke/krpc-go/lib/encode"
	service "github.com/atburke/krpc-go/lib/service"
	spacecenter "github.com/atburke/krpc-go/spacecenter"
	tracerr "github.com/ztrue/tracerr"
)

// Code generated by gen_services.go. DO NOT EDIT.

// Camera - a Docking Camera.
type Camera struct {
	service.BaseClass
}

// NewCamera creates a new Camera.
func NewCamera(id uint64, client *krpcgo.KRPCClient) *Camera {
	c := &Camera{BaseClass: service.BaseClass{Client: client}}
	c.SetID(id)
	return c
}

// DockingCamera - camera Service
type DockingCamera struct {
	Client *krpcgo.KRPCClient
}

// NewDockingCamera creates a new DockingCamera.
func NewDockingCamera(client *krpcgo.KRPCClient) *DockingCamera {
	return &DockingCamera{Client: client}
}

// Camera - get a Camera part
//
// Allowed game scenes: any.
func (s *DockingCamera) Camera(part *spacecenter.Part) (*Camera, error) {
	var err error
	var argBytes []byte
	var vv Camera
	request := &api.ProcedureCall{
		Procedure: "Camera",
		Service:   "DockingCamera",
	}
	argBytes, err = encode.Marshal(part)
	if err != nil {
		return &vv, tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x0),
		Value:    argBytes,
	})
	result, err := s.Client.Call(request)
	if err != nil {
		return &vv, tracerr.Wrap(err)
	}
	err = encode.Unmarshal(result.Value, &vv)
	if err != nil {
		return &vv, tracerr.Wrap(err)
	}
	vv.Client = s.Client
	return &vv, nil
}

// Available - check if the Camera API is avaiable
//
// Allowed game scenes: any.
func (s *DockingCamera) Available() (bool, error) {
	var err error
	var vv bool
	request := &api.ProcedureCall{
		Procedure: "get_Available",
		Service:   "DockingCamera",
	}
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

// StreamAvailable - check if the Camera API is avaiable
//
// Allowed game scenes: any.
func (s *DockingCamera) StreamAvailable() (*krpcgo.Stream[bool], error) {
	var err error
	request := &api.ProcedureCall{
		Procedure: "get_Available",
		Service:   "DockingCamera",
	}
	krpc := krpc.NewKRPC(s.Client)
	st, err := krpc.AddStream(request, true)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	rawStream := s.Client.GetStream(st.Id)
	stream := krpcgo.MapStream(rawStream, func(b []byte) bool {
		var value bool
		encode.Unmarshal(b, &value)
		return value
	})
	stream.AddCloser(func() error {
		return tracerr.Wrap(krpc.RemoveStream(st.Id))
	})
	return stream, nil
}

// Part - get the part containing this Camera.
//
// Allowed game scenes: any.
func (s *Camera) Part() (*spacecenter.Part, error) {
	var err error
	var argBytes []byte
	var vv spacecenter.Part
	request := &api.ProcedureCall{
		Procedure: "Camera_get_Part",
		Service:   "DockingCamera",
	}
	argBytes, err = encode.Marshal(s)
	if err != nil {
		return &vv, tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x0),
		Value:    argBytes,
	})
	result, err := s.Client.Call(request)
	if err != nil {
		return &vv, tracerr.Wrap(err)
	}
	err = encode.Unmarshal(result.Value, &vv)
	if err != nil {
		return &vv, tracerr.Wrap(err)
	}
	vv.Client = s.Client
	return &vv, nil
}

// Image - get the image.
//
// Allowed game scenes: any.
func (s *Camera) Image() ([]byte, error) {
	var err error
	var argBytes []byte
	var vv []byte
	request := &api.ProcedureCall{
		Procedure: "Camera_get_Image",
		Service:   "DockingCamera",
	}
	argBytes, err = encode.Marshal(s)
	if err != nil {
		return vv, tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x0),
		Value:    argBytes,
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

// StreamImage - get the image.
//
// Allowed game scenes: any.
func (s *Camera) StreamImage() (*krpcgo.Stream[[]byte], error) {
	var err error
	var argBytes []byte
	request := &api.ProcedureCall{
		Procedure: "Camera_get_Image",
		Service:   "DockingCamera",
	}
	argBytes, err = encode.Marshal(s)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	request.Arguments = append(request.Arguments, &api.Argument{
		Position: uint32(0x0),
		Value:    argBytes,
	})
	krpc := krpc.NewKRPC(s.Client)
	st, err := krpc.AddStream(request, true)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	rawStream := s.Client.GetStream(st.Id)
	stream := krpcgo.MapStream(rawStream, func(b []byte) []byte {
		var value []byte
		encode.Unmarshal(b, &value)
		return value
	})
	stream.AddCloser(func() error {
		return tracerr.Wrap(krpc.RemoveStream(st.Id))
	})
	return stream, nil
}
