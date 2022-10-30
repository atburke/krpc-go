package gen

import (
	"fmt"

	"github.com/atburke/krpc-go/api"
	"github.com/atburke/krpc-go/lib/utils"
	"github.com/dave/jennifer/jen"
	"github.com/mitchellh/go-wordwrap"
	"github.com/ztrue/tracerr"
)

const DocsLineLength = 77 // line length of 80 minus "// "

func wrap(s string) string {
	return wordwrap.WrapString(s, DocsLineLength)
}

func GenerateService(f *jen.File, service *api.Service) error {
	return nil
}

func GenerateProcedure(f *jen.File, procedure *api.Procedure) error {
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
	f.Comment(wrap(classDocs))
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
			jen.Comment(wrap(valueDocs)),
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
	f.Comment(wrap(docs))
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
		jen.Id("err").Id(exceptionName),
	).Id("Error").Params().String().Block(
		jen.Return(jen.Id("err").Dot("msg")),
	)

	return nil
}
