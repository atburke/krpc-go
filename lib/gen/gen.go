package gen

import (
	"fmt"
	"strings"

	"github.com/atburke/krpc-go/api"
	"github.com/atburke/krpc-go/lib/utils"
	"github.com/dave/jennifer/jen"
	"github.com/mitchellh/go-wordwrap"
	"github.com/ztrue/tracerr"
)

const DocsLineLength = 77 // line length of 80 minus "// "

func wrapDocComment(s string) string {
	wrapped := wordwrap.WrapString(s, DocsLineLength)
	inputLines := strings.Split(wrapped, "\n")
	var outputLines []string
	for _, line := range inputLines {
		outputLines = append(outputLines, strings.TrimSpace("// "+line))
	}
	return strings.Join(outputLines, "\n")
}

// GenerateService generates a service.
func GenerateService(f *jen.File, service *api.Service) error {
	for _, exception := range service.Exceptions {
		if err := GenerateException(f, exception); err != nil {
			return tracerr.Wrap(err)
		}
	}
	for _, enum := range service.Enumerations {
		if err := GenerateEnum(f, enum); err != nil {
			return tracerr.Wrap(err)
		}
	}
	for _, class := range service.Classes {
		if err := GenerateClass(f, class); err != nil {
			return tracerr.Wrap(err)
		}
	}

	serviceDocs, err := utils.ParseXMLDocumentation(service.Documentation, service.Name+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}

	f.Comment(wrapDocComment(serviceDocs))
	f.Type().Id(service.Name).Struct(
		jen.Id("Client").Op("*").Qual(clientMod, "KRPCClient"),
	)

	for _, procedure := range service.Procedures {
		if err := GenerateProcedure(f, service.Name, procedure); err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

func GenerateProcedure(f *jen.File, serviceName string, procedure *api.Procedure) error {
	var err error
	switch procedureType := GetProcedureType(procedure.Name); procedureType {
	case Procedure:
		err = generateProcedure(f, serviceName, procedure)
	case ServiceGetter:
		err = generateServiceGetter(f, serviceName, procedure)
	case ServiceSetter:
		err = generateServiceSetter(f, serviceName, procedure)
	case ClassMethod:
		err = generateClassMethod(f, serviceName, procedure)
	case StaticClassMethod:
		err = generateStaticClassMethod(f, serviceName, procedure)
	case ClassGetter:
		err = generateClassGetter(f, serviceName, procedure)
	case ClassSetter:
		err = generateClassSetter(f, serviceName, procedure)
	default:
		return tracerr.Errorf("Unknown procedure type: %v", procedureType)
	}
	return tracerr.Wrap(err)
}

func formatGameScenes(gameScenes []api.Procedure_GameScene) string {
	var scenes []string
	for _, scene := range gameScenes {
		scenes = append(scenes, scene.String())
	}
	var sceneString string
	if len(scenes) > 0 {
		sceneString = strings.Join(scenes, ", ")
	} else {
		sceneString = "any"
	}
	return fmt.Sprintf("Allowed game scenes: %v.", sceneString)
}

func generateProcedureBody(serviceName string, procedure *api.Procedure) (funcBody []jen.Code, params []jen.Code, returnType *jen.Statement) {
	returnType = GetGoType(procedure.ReturnType)

	// Define some variables
	funcBody = []jen.Code{
		jen.Var().Err().Error(),
	}

	if len(procedure.Parameters) > 0 {
		funcBody = append(funcBody,
			jen.Var().Id("argBytes").Index().Byte(),
		)
	}

	// Only create the return variable if needed
	if returnType != nil {
		funcBody = append(funcBody,
			jen.Var().Id("value").Add(returnType),
		)
	}

	// Define the request
	funcBody = append(funcBody,
		jen.Id("request").Op(":=").Op("&").Qual(apiMod, "ProcedureCall").Values(jen.Dict{
			jen.Id("Service"):   jen.Lit(serviceName),
			jen.Id("Procedure"): jen.Lit(procedure.Name),
		}),
	)

	// Shorthand for if err != nil {...
	var errReturn []jen.Code
	if returnType != nil {
		errReturn = []jen.Code{jen.Id("value"), jen.Qual(tracerrMod, "Wrap").Call(jen.Err())}
	} else {
		errReturn = []jen.Code{jen.Qual(tracerrMod, "Wrap").Call(jen.Err())}
	}
	errCheck := jen.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(errReturn...),
	)

	// Marshal arguments
	_, err := GetClassName(procedure.Name)
	isClass := err == nil
	for i, param := range procedure.Parameters {
		// If this is any kind of class method, use the class itself as the first param
		if i == 0 && isClass {
			param.Name = "s"
		} else {
			paramType := GetGoType(param.Type)
			params = append(params, jen.Id(param.Name).Add(paramType))
		}

		funcBody = append(funcBody,
			jen.List(jen.Id("argBytes"), jen.Err()).Op("=").Qual(encodeMod, "Marshal").Call(
				jen.Id(param.Name),
			),
			errCheck,
			jen.Id("request").Dot("Arguments").Op("=").Append(
				jen.Id("request").Dot("Arguments"),
				jen.Op("&").Qual(apiMod, "Argument").Values(jen.Dict{
					jen.Id("Position"): jen.Lit(uint32(i)),
					jen.Id("Value"):    jen.Id("argBytes"),
				}),
			),
		)
	}

	// Call the procedure
	funcBody = append(funcBody,
		jen.List(
			jen.Id("result"), jen.Err(),
		).Op(":=").Id("s").Dot("Client").Dot("Call").Call(
			jen.Id("request"), jen.Lit(returnType != nil),
		),
		errCheck,
	)

	if returnType != nil {
		// Unmarshal the result bytes
		funcBody = append(funcBody,
			jen.Err().Op("=").Qual(encodeMod, "Unmarshal").Call(jen.Id("result"), jen.Op("&").Id("value")),
			errCheck,
			jen.Return(jen.Id("value"), jen.Nil()),
		)
	} else {
		funcBody = append(funcBody,
			jen.Return(jen.Nil()),
		)
	}

	// Return captured variables
	return
}

func generateBaseProcedure(f *jen.File, procName, procDocs, receiver, serviceName string, procedure *api.Procedure) {
	funcBody, params, returnType := generateProcedureBody(serviceName, procedure)

	var retType jen.Code
	if returnType != nil {
		retType = jen.Parens(jen.List(returnType, jen.Error()))
	} else {
		retType = jen.Error()
	}
	// Define the procedure function
	f.Comment(wrapDocComment(procDocs))
	f.Func().Params(
		jen.Id("s").Op("*").Id(receiver),
	).Id(procName).Params(params...).Add(retType).Block(funcBody...)
}

func generateProcedure(f *jen.File, serviceName string, procedure *api.Procedure) error {
	procName := procedure.Name
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" will ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, serviceName, serviceName, procedure)

	return nil
}

func generateServiceGetter(f *jen.File, serviceName string, procedure *api.Procedure) error {
	propName, err := GetPropertyName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := propName
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" will ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, serviceName, serviceName, procedure)

	return nil
}

func generateServiceSetter(f *jen.File, serviceName string, procedure *api.Procedure) error {
	propName, err := GetPropertyName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := "Set" + propName
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" will ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, serviceName, serviceName, procedure)

	return nil
}

func generateClassMethod(f *jen.File, serviceName string, procedure *api.Procedure) error {
	className, err := GetClassName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := GetProcedureName(procedure.Name)
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" will ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, className, serviceName, procedure)

	return nil
}

func generateStaticClassMethod(f *jen.File, serviceName string, procedure *api.Procedure) error {
	return tracerr.Wrap(generateClassMethod(f, serviceName, procedure))
}

func generateClassGetter(f *jen.File, serviceName string, procedure *api.Procedure) error {
	className, err := GetClassName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	propName, err := GetPropertyName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := propName
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" will ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, className, serviceName, procedure)

	return nil
}

func generateClassSetter(f *jen.File, serviceName string, procedure *api.Procedure) error {
	className, err := GetClassName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	propName, err := GetPropertyName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := "Set" + propName
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" will ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, className, serviceName, procedure)

	return nil
}

// GenerateClass generates a struct for a given class definition.
func GenerateClass(f *jen.File, class *api.Class) error {
	className := class.Name
	classDocs, err := utils.ParseXMLDocumentation(class.Documentation, className+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}

	// Define the struct.
	f.Comment(wrapDocComment(classDocs))
	f.Type().Id(className).Struct(
		jen.Id("BaseClass"),
	)

	// Define the constructor.
	constructorName := "New" + className
	f.Comment(fmt.Sprintf("%v creates a new %v.", constructorName, className))
	f.Func().Id(constructorName).Params(
		jen.Id("id").Uint64(),
		jen.Id("client").Op("*").Qual("github.com/atburke/krpc-go/lib/client", "KRPCClient"),
	).Op("*").Id(className).Block(
		jen.Return(jen.Op("&").Id(className).Values(jen.Dict{
			jen.Id("ID"):     jen.Id("id"),
			jen.Id("Client"): jen.Id("client"),
		})),
	)
	return nil
}

// GenerateEnum generates an enum for a given enum definition.
func GenerateEnum(f *jen.File, enum *api.Enumeration) error {
	enumName := enum.Name
	enumDocs, err := utils.ParseXMLDocumentation(enum.Documentation, enumName+" is ")
	if err != nil {
		return tracerr.Wrap(err)
	}

	// Define the enum type.
	f.Comment(wordwrap.WrapString(enumDocs, DocsLineLength))
	f.Type().Id(enumName).Int32()

	// Define the enum values.
	var defs []jen.Code
	for _, value := range enum.Values {
		valueName := fmt.Sprintf("%v_%v", enumName, value.Name)
		valueDocs, err := utils.ParseXMLDocumentation(value.Documentation, "")
		if err != nil {
			return tracerr.Wrap(err)
		}
		defs = append(defs,
			jen.Comment(wrapDocComment(valueDocs)),
			jen.Id(valueName).Id(enumName).Op("=").Lit(value.Value),
		)
	}

	f.Const().Defs(defs...)
	return nil
}

// GenerateException generates an error for a given exception definition.
func GenerateException(f *jen.File, exception *api.Exception) error {
	// Names are given in the format XYZException. We want the more go-like
	// ErrXYZ.
	exceptionName := "Err" + exception.Name[:len(exception.Name)-len("exception")]
	docs, err := utils.ParseXMLDocumentation(exception.Documentation, exceptionName+" means ")
	if err != nil {
		return tracerr.Wrap(err)
	}

	// Define the error type.
	f.Comment(wrapDocComment(docs))
	f.Type().Id(exceptionName).Struct(
		jen.Id("msg").String(),
	)

	// Define the constructor.
	constructorName := "New" + exceptionName
	f.Comment(fmt.Sprintf("%v creates a new %v.", constructorName, exceptionName))
	f.Func().Id(constructorName).Params(
		jen.Id("msg").String(),
	).Op("*").Id(exceptionName).Block(
		jen.Return(jen.Op("&").Id(exceptionName).Values(jen.Dict{
			jen.Id("msg"): jen.Id("msg"),
		})),
	)

	// Define the Error() function.
	f.Comment("Error returns a human-readable error.")
	f.Func().Params(
		jen.Err().Id(exceptionName),
	).Id("Error").Params().String().Block(
		jen.Return(jen.Err().Dot("msg")),
	)

	return nil
}
