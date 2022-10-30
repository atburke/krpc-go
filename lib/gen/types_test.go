package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"testing"

	"github.com/atburke/krpc-go/api"
	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
)

func TestGetProcedureType(t *testing.T) {
	tests := []struct {
		input        string
		expectedType ProcedureType
	}{
		{
			input:        "SomeProcedure",
			expectedType: Procedure,
		},
		{
			input:        "get_MyProperty",
			expectedType: ServiceGetter,
		},
		{
			input:        "set_MyProperty",
			expectedType: ServiceSetter,
		},
		{
			input:        "MyClass_MyMethod",
			expectedType: ClassMethod,
		},
		{
			input:        "MyClass_static_MyMethod",
			expectedType: StaticClassMethod,
		},
		{
			input:        "MyClass_get_MyProperty",
			expectedType: ClassGetter,
		},
		{
			input:        "MyClass_set_MyProperty",
			expectedType: ClassSetter,
		},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			require.Equal(t, tc.expectedType, GetProcedureType(tc.input))
		})
	}
}

func TestGetPropertyName(t *testing.T) {
	tests := []struct {
		input          string
		checkErr       require.ErrorAssertionFunc
		expectedOutput string
	}{
		{
			input:          "set_MyProperty",
			checkErr:       require.NoError,
			expectedOutput: "MyProperty",
		},
		{
			input:          "MyClass_get_MyProperty",
			checkErr:       require.NoError,
			expectedOutput: "MyProperty",
		},
		{
			input:    "MyClass_MyMethod",
			checkErr: require.Error,
		},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			out, err := GetPropertyName(tc.input)
			tc.checkErr(t, err)
			require.Equal(t, tc.expectedOutput, out)
		})
	}
}

func TestGetClassName(t *testing.T) {
	tests := []struct {
		input          string
		checkErr       require.ErrorAssertionFunc
		expectedOutput string
	}{
		{
			input:          "MyClass_MyMethod",
			checkErr:       require.NoError,
			expectedOutput: "MyClass",
		},
		{
			input:          "MyClass_set_MyProperty",
			checkErr:       require.NoError,
			expectedOutput: "MyClass",
		},
		{
			input:    "get_MyProperty",
			checkErr: require.Error,
		},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			out, err := GetClassName(tc.input)
			tc.checkErr(t, err)
			require.Equal(t, tc.expectedOutput, out)
		})
	}
}

func TestGetGoType(t *testing.T) {
	tests := []struct {
		name         string
		t            *api.Type
		wantAPI      bool
		wantService  bool
		expectedType string
	}{
		{
			name: "primitive",
			t: &api.Type{
				Code: api.Type_UINT64,
			},
			expectedType: "uint64",
		},
		{
			name: "class",
			t: &api.Type{
				Code: api.Type_CLASS,
				Name: "MyClass",
			},
			wantService:  true,
			expectedType: "service.MyClass",
		},
		{
			name: "special",
			t: &api.Type{
				Code: api.Type_PROCEDURE_CALL,
			},
			wantAPI:      true,
			expectedType: "api.ProcedureCall",
		},
		{
			name: "tuple",
			t: &api.Type{
				Code: api.Type_TUPLE,
				Types: []*api.Type{
					{
						Code: api.Type_STRING,
					},
					{
						Code: api.Type_BOOL,
					},
					{
						Code: api.Type_DOUBLE,
					},
				},
			},
			wantAPI:      true,
			expectedType: "api.Tuple3[string, bool, float64]",
		},
		{
			name: "list",
			t: &api.Type{
				Code: api.Type_LIST,
				Types: []*api.Type{
					{
						Code: api.Type_SINT64,
					},
				},
			},
			expectedType: "[]int64",
		},
		{
			name: "set",
			t: &api.Type{
				Code: api.Type_SET,
				Types: []*api.Type{
					{
						Code: api.Type_STRING,
					},
				},
			},
			expectedType: "map[string]struct{}",
		},
		{
			name: "dictionary",
			t: &api.Type{
				Code: api.Type_DICTIONARY,
				Types: []*api.Type{
					{
						Code: api.Type_STRING,
					},
					{
						Code: api.Type_FLOAT,
					},
				},
			},
			expectedType: "map[string]float32",
		},
	}
	for _, tc := range tests {
		t.Run(tc.expectedType, func(t *testing.T) {
			var imports []string
			if tc.wantAPI {
				imports = append(imports, `import api "github.com/atburke/krpc-go/api"`)
			}
			if tc.wantService {
				imports = append(imports, `import service "github.com/atburke/krpc-go/lib/service"`)
			}
			expectedRaw := fmt.Sprintf(`
			package gentest

			%v

			type Test %v
			`, strings.Join(imports, "\n"), tc.expectedType)
			expectedOut, err := format.Source([]byte(expectedRaw))
			require.NoError(t, err)

			f := jen.NewFile("gentest")
			f.Type().Id("Test").Add(GetGoType(tc.t))
			var out bytes.Buffer
			require.NoError(t, f.Render(&out))
			require.Equal(t, string(expectedOut), out.String())
		})
	}
}
