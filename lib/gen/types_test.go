package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"testing"

	"github.com/atburke/krpc-go/types"
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
		t            *types.Type
		wantAPI      bool
		wantService  string
		expectedType string
	}{
		{
			name: "primitive",
			t: &types.Type{
				Code: types.Type_UINT64,
			},
			expectedType: "uint64",
		},
		{
			name: "class",
			t: &types.Type{
				Code:    types.Type_CLASS,
				Name:    "MyClass",
				Service: "MyService",
			},
			expectedType: "*MyClass",
		},
		{
			name: "class from another package",
			t: &types.Type{
				Code:    types.Type_CLASS,
				Name:    "MyClass",
				Service: "MyOtherService",
			},
			wantService:  "MyOtherService",
			expectedType: "*myotherservice.MyClass",
		},
		{
			name: "special",
			t: &types.Type{
				Code: types.Type_PROCEDURE_CALL,
			},
			wantAPI:      true,
			expectedType: "*types.ProcedureCall",
		},
		{
			name: "tuple",
			t: &types.Type{
				Code: types.Type_TUPLE,
				Types: []*types.Type{
					{
						Code: types.Type_STRING,
					},
					{
						Code: types.Type_BOOL,
					},
					{
						Code: types.Type_DOUBLE,
					},
				},
			},
			wantAPI:      true,
			expectedType: "types.Tuple3[string, bool, float64]",
		},
		{
			name: "list",
			t: &types.Type{
				Code: types.Type_LIST,
				Types: []*types.Type{
					{
						Code: types.Type_SINT64,
					},
				},
			},
			expectedType: "[]int64",
		},
		{
			name: "set",
			t: &types.Type{
				Code: types.Type_SET,
				Types: []*types.Type{
					{
						Code: types.Type_STRING,
					},
				},
			},
			expectedType: "map[string]struct{}",
		},
		{
			name: "dictionary",
			t: &types.Type{
				Code: types.Type_DICTIONARY,
				Types: []*types.Type{
					{
						Code: types.Type_STRING,
					},
					{
						Code: types.Type_FLOAT,
					},
				},
			},
			expectedType: "map[string]float32",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var imports []string
			if tc.wantAPI {
				imports = append(imports, `import types "github.com/atburke/krpc-go/types"`)
			}
			if tc.wantService != "" {
				imports = append(imports, fmt.Sprintf("import %v %q", strings.ToLower(tc.wantService), getServicePackage(tc.wantService)))
			}
			expectedRaw := fmt.Sprintf(`
			package gentest

			%v

			type Test %v
			`, strings.Join(imports, "\n"), tc.expectedType)
			expectedOut, err := format.Source([]byte(expectedRaw))
			require.NoError(t, err)

			f := jen.NewFile("gentest")
			f.Type().Id("Test").Add(GetGoType(tc.t, WithPackage("github.com/atburke/krpc-go/myservice")))
			var out bytes.Buffer
			require.NoError(t, f.Render(&out))
			require.Equal(t, string(expectedOut), out.String())
		})
	}
}
