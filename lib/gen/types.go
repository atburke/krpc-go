package gen

import (
	"fmt"
	"regexp"

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
