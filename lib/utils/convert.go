package utils

import (
	"github.com/golang/protobuf/proto"
	"github.com/ztrue/tracerr"
)

func EncodeBool(b bool) []byte {
	if b {
		return proto.EncodeVarint(1)
	}
	return proto.EncodeVarint(0)
}

func DecodeBool(d []byte) (bool, error) {
	b, size := proto.DecodeVarint(d)
	if size == 0 {
		return false, tracerr.Errorf("Failed to decode bool: %v", d)
	}
	return b != 0, nil
}
