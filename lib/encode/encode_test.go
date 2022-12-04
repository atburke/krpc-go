package encode

import (
	"reflect"
	"testing"

	"github.com/atburke/krpc-go/lib/service"
	"github.com/atburke/krpc-go/types"
	"github.com/stretchr/testify/require"
)

type testClass struct {
	service.BaseClass
}

func newTestClass(id uint64) *testClass {
	c := testClass{}
	c.SetID(id)
	return &c
}

type testEnum int32

const (
	a testEnum = 1
	b testEnum = 2
)

func (v testEnum) Value() int32 {
	return int32(v)
}

func (v *testEnum) SetValue(value int32) {
	*v = testEnum(value)
}

func TestMarshalAndUnmarshal(t *testing.T) {
	tests := []struct {
		name              string
		input             interface{}
		skipOutputPointer bool
	}{
		{
			name:  "uint64",
			input: uint64(99),
		},
		{
			name:  "int32",
			input: int32(-43),
		},
		{
			name:  "bool",
			input: true,
		},
		{
			name:  "float64",
			input: float64(3.14159265),
		},
		{
			name:  "string",
			input: "hello there!",
		},
		{
			name:              "class",
			input:             newTestClass(99),
			skipOutputPointer: true,
		},
		{
			name:  "enum",
			input: a,
		},
		{
			name:  "slice",
			input: []string{"test1", "test2", "test3"},
		},
		{
			name: "set",
			input: map[uint64]struct{}{
				66: {},
				99: {},
				5:  {},
			},
		},
		{
			name: "map",
			input: map[string]int64{
				"a": -1,
				"b": 2,
				"c": -9999,
			},
		},
		{
			name:  "tuple",
			input: types.NewTuple3("test", uint64(77), float64(6.28)),
		},
		{
			name:  "slice of pointers",
			input: []*testClass{newTestClass(1), newTestClass(2)},
		},
		{
			name:  "map of pointers",
			input: map[string]*testClass{"1": newTestClass(1), "2": newTestClass(2)},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var output interface{}
			if tc.skipOutputPointer {
				output = reflect.New(reflect.TypeOf(tc.input).Elem()).Interface()
			} else {
				output = reflect.New(reflect.TypeOf(tc.input)).Interface()
			}

			b, err := Marshal(tc.input)
			require.NoError(t, err)
			err = Unmarshal(b, output)
			require.NoError(t, err)
			if !tc.skipOutputPointer {
				output = reflect.ValueOf(output).Elem().Interface()
			}

			require.Equal(t, tc.input, output)
		})
	}
}
