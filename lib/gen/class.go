package gen

import (
	"fmt"

	"github.com/atburke/krpc-go/api"
	"github.com/atburke/krpc-go/lib/utils"
	"github.com/dave/jennifer/jen"
	"github.com/ztrue/tracerr"
)

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
