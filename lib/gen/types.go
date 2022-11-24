package gen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/atburke/krpc-go/api"
	"github.com/dave/jennifer/jen"
	"github.com/ztrue/tracerr"
)

// ProcedureType is the type of a procedure.
type ProcedureType int

const (
	// Procedure is just part of the service.
	Procedure ProcedureType = iota
	// Procedure gets a property of the service.
	ServiceGetter
	// Procedure sets a property of the service.
	ServiceSetter
	// A class method.
	ClassMethod
	// A static class method.
	StaticClassMethod
	// A class property getter.
	ClassGetter
	// A class property setter.
	ClassSetter
)

const procID = "([a-zA-Z0-9]+)"

var serviceGetterRE = regexp.MustCompile(fmt.Sprintf(`^get_%v$`, procID))
var serviceSetterRE = regexp.MustCompile(fmt.Sprintf(`^set_%v$`, procID))
var classMethodRE = regexp.MustCompile(fmt.Sprintf(`^%v_%v$`, procID, procID))
var staticClassMethodRE = regexp.MustCompile(fmt.Sprintf(`^%v_static_%v$`, procID, procID))
var classGetterRE = regexp.MustCompile(fmt.Sprintf(`^%v_get_%v$`, procID, procID))
var classSetterRE = regexp.MustCompile(fmt.Sprintf(`^%v_set_%v$`, procID, procID))

// GetProcedureType determines the type of a procedure from its name.
func GetProcedureType(procedureName string) ProcedureType {
	switch {
	case staticClassMethodRE.MatchString(procedureName):
		return StaticClassMethod
	case classGetterRE.MatchString(procedureName):
		return ClassGetter
	case classSetterRE.MatchString(procedureName):
		return ClassSetter
	case serviceGetterRE.MatchString(procedureName):
		return ServiceGetter
	case serviceSetterRE.MatchString(procedureName):
		return ServiceSetter
	case classMethodRE.MatchString(procedureName):
		return ClassMethod
	default:
		return Procedure
	}
}

// GetPropertyName gets the name of a property from a procedure's name. Returns
// an error if the procedure is not for a property.
func GetPropertyName(procedureName string) (string, error) {
	switch GetProcedureType(procedureName) {
	case ServiceGetter:
		return serviceGetterRE.FindStringSubmatch(procedureName)[1], nil
	case ServiceSetter:
		return serviceSetterRE.FindStringSubmatch(procedureName)[1], nil
	case ClassGetter:
		return classGetterRE.FindStringSubmatch(procedureName)[2], nil
	case ClassSetter:
		return classSetterRE.FindStringSubmatch(procedureName)[2], nil
	default:
		return "", tracerr.Errorf("Procedure %q does not have a property", procedureName)
	}
}

// GetClassName gets the name of a class from a procedure's name. Returns an
// error if the procedure is not for a class.
func GetClassName(procedureName string) (string, error) {
	switch GetProcedureType(procedureName) {
	case ClassMethod:
		return classMethodRE.FindStringSubmatch(procedureName)[1], nil
	case StaticClassMethod:
		return staticClassMethodRE.FindStringSubmatch(procedureName)[1], nil
	case ClassGetter:
		return classGetterRE.FindStringSubmatch(procedureName)[1], nil
	case ClassSetter:
		return classSetterRE.FindStringSubmatch(procedureName)[1], nil
	default:
		return "", tracerr.Errorf("Procedure %q does not have a class", procedureName)
	}
}

// GetProcedureName gets the name of a procedure.
func GetProcedureName(procedureName string) string {
	terms := strings.Split(procedureName, "_")
	return terms[len(terms)-1]
}

// GetGoType gets the Go representation of a kRPC type.
func GetGoType(t *api.Type, pkg string) *jen.Statement {
	if t == nil {
		return nil
	}

	switch t.Code {
	// Special KRPC types.
	case api.Type_PROCEDURE_CALL:
		return jen.Qual(apiPkg, "ProcedureCall")
	case api.Type_STREAM:
		return jen.Qual(apiPkg, "Stream")
	case api.Type_STATUS:
		return jen.Qual(apiPkg, "Status")
	case api.Type_SERVICES:
		return jen.Qual(apiPkg, "Services")

	// Class or enum.
	case api.Type_CLASS, api.Type_ENUMERATION:
		if p := getServicePackage(t.Service); p == pkg {
			return jen.Id(t.Name)
		} else {
			return jen.Qual(p, t.Name)
		}

	// Primitives.
	case api.Type_DOUBLE:
		return jen.Float64()
	case api.Type_FLOAT:
		return jen.Float32()
	case api.Type_SINT32:
		return jen.Int32()
	case api.Type_SINT64:
		return jen.Int64()
	case api.Type_UINT32:
		return jen.Uint32()
	case api.Type_UINT64:
		return jen.Uint64()
	case api.Type_BOOL:
		return jen.Bool()
	case api.Type_STRING:
		return jen.String()
	case api.Type_BYTES:
		return jen.Index().Byte()

	// Collections.
	case api.Type_TUPLE:
		var tupleTypes []jen.Code
		for _, subType := range t.Types {
			tupleTypes = append(tupleTypes, GetGoType(subType, pkg))
		}
		return jen.Qual(
			apiPkg, fmt.Sprintf("Tuple%v", len(t.Types)),
		).Types(tupleTypes...)

	case api.Type_LIST:
		return jen.Index().Add(GetGoType(t.Types[0], pkg))
	case api.Type_SET:
		return jen.Map(GetGoType(t.Types[0], pkg)).Struct()
	case api.Type_DICTIONARY:
		return jen.Map(GetGoType(t.Types[0], pkg)).Add(GetGoType(t.Types[1], pkg))
	}

	// Type is None or unrecognized.
	return nil
}
