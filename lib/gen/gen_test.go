package gen

import (
	"bytes"
	"go/format"
	"testing"

	"github.com/atburke/krpc-go/api"
	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
)

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
