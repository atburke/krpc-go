package krpcgo

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func FuzzReadMessageLength(f *testing.F) {
	tests := []uint64{0, 1, 564, 9999999999999999934}
	for _, tc := range tests {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, i uint64) {
		data := proto.EncodeVarint(i)
		l, err := readMessageLength(bytes.NewReader(data))
		require.NoError(t, err)
		require.Equal(t, i, l)
	})
}
