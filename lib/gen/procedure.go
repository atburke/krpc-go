package gen

import (
	"fmt"
	"strings"

	"github.com/atburke/krpc-go/lib/utils"
	"github.com/atburke/krpc-go/types"
	"github.com/dave/jennifer/jen"
	"github.com/ztrue/tracerr"
)

// GenerateProcedure generates a procedure function from a given procedure definition.
func GenerateProcedure(f *jen.File, serviceName string, procedure *types.Procedure) error {
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

// formatGameScences formats a list of allowed game scenes as a string.
func formatGameScenes(gameScenes []types.Procedure_GameScene) string {
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

// generateProcedureBody generates the function body for a procedure.
func generateProcedureBody(serviceName string, procedure *types.Procedure) (funcBody []jen.Code, params []jen.Code, returnType *jen.Statement) {
	pkg := getServicePackage(serviceName)
	returnType = GetGoType(procedure.ReturnType, WithPackage(pkg))
	retVarType := GetGoType(procedure.ReturnType, WithPackage(pkg), NoPointerForClass)

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
			jen.Var().Id("vv").Add(retVarType),
		)
	}

	// Define the request
	funcBody = append(funcBody,
		jen.Id("request").Op(":=").Op("&").Qual(typesPkg, "ProcedureCall").Values(jen.Dict{
			jen.Id("Service"):   jen.Lit(serviceName),
			jen.Id("Procedure"): jen.Lit(procedure.Name),
		}),
	)

	// Shorthand for if err != nil {...
	var errReturn []jen.Code
	var returnVar *jen.Statement
	if returnType != nil {
		if isPointerType(procedure.ReturnType.Code) {
			returnVar = jen.Op("&").Id("vv")
		} else {
			returnVar = jen.Id("vv")
		}
		errReturn = []jen.Code{returnVar, jen.Qual(tracerrPkg, "Wrap").Call(jen.Err())}
	} else {
		errReturn = []jen.Code{jen.Qual(tracerrPkg, "Wrap").Call(jen.Err())}
	}
	errCheck := jen.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(errReturn...),
	)

	// Marshal arguments
	_, err := GetClassName(procedure.Name)
	isClass := err == nil
	for i, param := range procedure.Parameters {
		param.Name = utils.SanitizeIdentifier(param.Name)
		// If this is any kind of class method, use the class itself as the first param
		if i == 0 && isClass {
			param.Name = "s"
		} else {
			paramType := GetGoType(param.Type, WithPackage(pkg))
			params = append(params, jen.Id(param.Name).Add(paramType))
		}

		funcBody = append(funcBody,
			jen.List(jen.Id("argBytes"), jen.Err()).Op("=").Qual(encodePkg, "Marshal").Call(
				jen.Id(param.Name),
			),
			errCheck,
			jen.Id("request").Dot("Arguments").Op("=").Append(
				jen.Id("request").Dot("Arguments"),
				jen.Op("&").Qual(typesPkg, "Argument").Values(jen.Dict{
					jen.Id("Position"): jen.Lit(uint32(i)),
					jen.Id("Value"):    jen.Id("argBytes"),
				}),
			),
		)
	}

	// Call the procedure
	var lhs *jen.Statement
	if returnType != nil {
		lhs = jen.List(
			jen.Id("result"), jen.Err(),
		).Op(":=")
	} else {
		lhs = jen.List(
			jen.Id("_"), jen.Err(),
		).Op("=")
	}
	funcBody = append(funcBody,
		lhs.Id("s").Dot("Client").Dot("Call").Call(
			jen.Id("request"),
		),
		errCheck,
	)

	if returnType != nil {
		// Unmarshal the result bytes
		funcBody = append(funcBody,
			jen.Err().Op("=").Qual(encodePkg, "Unmarshal").Call(jen.Id("result").Dot("Value"), jen.Op("&").Id("vv")),
			errCheck,
		)
		if procedure.ReturnType.Code == types.Type_CLASS {
			funcBody = append(funcBody,
				jen.Id("vv").Dot("Client").Op("=").Id("s").Dot("Client"),
			)
		}
		funcBody = append(funcBody,
			jen.Return(returnVar, jen.Nil()),
		)
	} else {
		funcBody = append(funcBody,
			jen.Return(jen.Nil()),
		)
	}

	// Return captured variables
	return
}

// generateBaseProcedure generates a procedure function using extra info about the call signature.
func generateBaseProcedure(f *jen.File, procName, procDocs, receiver, serviceName string, procedure *types.Procedure) {
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

	// If this procedure has a return value, also generate a stream definition
	// Note: not streaming classes for simplicity, may change later
	if returnType != nil && !isPointerType(procedure.ReturnType.Code) {
		funcBody, streamRetType := generateStreamBody(serviceName, procedure)
		streamFuncName := procName + "Stream"
		f.Comment(wrapDocComment(strings.ReplaceAll(procDocs, procName, streamFuncName)))
		f.Func().Params(
			jen.Id("s").Op("*").Id(receiver),
		).Id(streamFuncName).Params(params...).Add(jen.Parens(jen.List(streamRetType, jen.Error()))).Block(funcBody...)
	}
}

func generateStreamBody(serviceName string, procedure *types.Procedure) (funcBody []jen.Code, returnType *jen.Statement) {
	internalReturnType := GetGoType(procedure.ReturnType, WithPackage(getServicePackage(serviceName)))
	returnType = jen.Op("*").Qual(krpcPkg, "Stream").Types(internalReturnType)

	funcBody = []jen.Code{
		jen.Var().Err().Error(),
	}

	if len(procedure.Parameters) > 0 {
		funcBody = append(funcBody,
			jen.Var().Id("argBytes").Index().Byte(),
		)
	}

	funcBody = append(funcBody,
		jen.Id("request").Op(":=").Op("&").Qual(typesPkg, "ProcedureCall").Values(jen.Dict{
			jen.Id("Service"):   jen.Lit(serviceName),
			jen.Id("Procedure"): jen.Lit(procedure.Name),
		}),
	)

	// Shorthand for if err != nil {...
	errReturn := []jen.Code{jen.Nil(), jen.Qual(tracerrPkg, "Wrap").Call(jen.Err())}
	errCheck := jen.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(errReturn...),
	)

	// Marshal arguments
	_, err := GetClassName(procedure.Name)
	isClass := err == nil
	for i, param := range procedure.Parameters {
		param.Name = utils.SanitizeIdentifier(param.Name)
		// If this is any kind of class method, use the class itself as the first param
		if i == 0 && isClass {
			param.Name = "s"
		}

		funcBody = append(funcBody,
			jen.List(jen.Id("argBytes"), jen.Err()).Op("=").Qual(encodePkg, "Marshal").Call(
				jen.Id(param.Name),
			),
			errCheck,
			jen.Id("request").Dot("Arguments").Op("=").Append(
				jen.Id("request").Dot("Arguments"),
				jen.Op("&").Qual(typesPkg, "Argument").Values(jen.Dict{
					jen.Id("Position"): jen.Lit(uint32(i)),
					jen.Id("Value"):    jen.Id("argBytes"),
				}),
			),
		)
	}

	var krpcConstructor *jen.Statement
	if serviceName == "KRPC" {
		krpcConstructor = jen.Id("New")
	} else {
		krpcConstructor = jen.Qual(getServicePackage("KRPC"), "New")
	}

	funcBody = append(funcBody,
		jen.Id("krpc").Op(":=").Add(krpcConstructor).Call(jen.Id("s").Dot("Client")),

		// Start the stream
		jen.List(jen.Id("st"), jen.Err()).Op(":=").Id("krpc").Dot("AddStream").Call(
			jen.Id("request"), jen.Lit(true),
		),
		errCheck,

		jen.Id("rawStream").Op(":=").Id("s").Dot("Client").Dot("GetStream").Call(
			jen.Id("st").Dot("Id"),
		),

		jen.Id("stream").Op(":=").Qual(krpcPkg, "MapStream").Call(
			jen.Id("rawStream"),
			jen.Func().Params(jen.Id("b").Index().Byte()).Add(internalReturnType).Block(
				jen.Var().Id("value").Add(internalReturnType),
				jen.Qual(encodePkg, "Unmarshal").Call(jen.Id("b"), jen.Op("&").Id("value")),
				jen.Return(jen.Id("value")),
			),
		),
		jen.Id("stream").Dot("AddCloser").Call(jen.Func().Params().Error().Block(
			jen.Return(jen.Qual(tracerrPkg, "Wrap").Call(
				jen.Id("krpc").Dot("RemoveStream").Call(jen.Id("st").Dot("Id")),
			)),
		)),
		jen.Return(jen.Id("stream"), jen.Nil()),
	)

	return
}

func generateProcedure(f *jen.File, serviceName string, procedure *types.Procedure) error {
	procName := procedure.Name
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, serviceName, serviceName, procedure)

	return nil
}

func generateServiceGetter(f *jen.File, serviceName string, procedure *types.Procedure) error {
	propName, err := GetPropertyName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := propName
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, serviceName, serviceName, procedure)

	return nil
}

func generateServiceSetter(f *jen.File, serviceName string, procedure *types.Procedure) error {
	propName, err := GetPropertyName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := "Set" + propName
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, serviceName, serviceName, procedure)

	return nil
}

func generateClassMethod(f *jen.File, serviceName string, procedure *types.Procedure) error {
	className, err := GetClassName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := GetProcedureName(procedure.Name)
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, className, serviceName, procedure)

	return nil
}

func generateStaticClassMethod(f *jen.File, serviceName string, procedure *types.Procedure) error {
	return tracerr.Wrap(generateClassMethod(f, serviceName, procedure))
}

func generateClassGetter(f *jen.File, serviceName string, procedure *types.Procedure) error {
	className, err := GetClassName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	propName, err := GetPropertyName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := propName
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, className, serviceName, procedure)

	return nil
}

func generateClassSetter(f *jen.File, serviceName string, procedure *types.Procedure) error {
	className, err := GetClassName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	propName, err := GetPropertyName(procedure.Name)
	if err != nil {
		return tracerr.Wrap(err)
	}
	procName := "Set" + propName
	procDocs, err := utils.ParseXMLDocumentation(procedure.Documentation, procName+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	procDocs = fmt.Sprintf("%v\n\n%v", procDocs, formatGameScenes(procedure.GameScenes))
	generateBaseProcedure(f, procName, procDocs, className, serviceName, procedure)

	return nil
}
