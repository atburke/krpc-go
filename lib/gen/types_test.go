package gen

import (
	"testing"

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
