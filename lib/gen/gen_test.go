package gen

import (
	"bytes"
	"go/format"
	"testing"

	"github.com/atburke/krpc-go/api"
	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
)

const testEnum = `
package gentest


// Test is a test enum.
type Test int32

const (
	// The first enum value.
	Test_One Test = int32(1)
	// The second enum value.
	Test_Two Test = int32(2)
	// The third enum value.
	Test_Three Test = int32(3)
)
`

func TestGenerateEnum(t *testing.T) {
	expectedOut, err := format.Source([]byte(testEnum))
	require.NoError(t, err)

	enum := &api.Enumeration{
		Name:          "Test",
		Documentation: "<summary>A test enum.</summary>",
		Values: []*api.EnumerationValue{
			{
				Name:          "One",
				Value:         1,
				Documentation: "<summary>The first enum value.</summary>",
			},
			{
				Name:          "Two",
				Value:         2,
				Documentation: "<summary>The second enum value.</summary>",
			},
			{
				Name:          "Three",
				Value:         3,
				Documentation: "<summary>The third enum value.</summary>",
			},
		},
	}
	f := jen.NewFile("gentest")
	require.NoError(t, GenerateEnum(f, enum))

	var out bytes.Buffer
	require.NoError(t, f.Render(&out))
	require.Equal(t, string(expectedOut), out.String())
}

const testException = `
package gentest


// ErrTest means the exception generating code is being tested.
type ErrTest struct {
	msg string
}

// NewErrTest creates a new ErrTest.
func NewErrTest(msg string) *ErrTest {
	return &ErrTest{msg: msg}
}

// Error returns a human-readable error.
func (err ErrTest) Error() string {
	return err.msg
}
`

func TestGenerateException(t *testing.T) {
	expectedOut, err := format.Source([]byte(testException))
	require.NoError(t, err)

	exception := &api.Exception{
		Name:          "TestException",
		Documentation: "<summary>The exception generating code is being tested.</summary>",
	}
	f := jen.NewFile("gentest")
	require.NoError(t, GenerateException(f, exception))

	var out bytes.Buffer
	require.NoError(t, f.Render(&out))
	require.Equal(t, string(expectedOut), out.String())
}
