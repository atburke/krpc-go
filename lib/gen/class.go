package gen

import (
	"fmt"

	"github.com/atburke/krpc-go/lib/utils"
	"github.com/atburke/krpc-go/types"
	"github.com/dave/jennifer/jen"
	"github.com/ztrue/tracerr"
)

// GenerateClass generates a struct for a given class definition.
func GenerateClass(f *jen.File, class *types.Class) error {
	className := class.Name
	classDocs, err := utils.ParseXMLDocumentation(class.Documentation, className+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}

	// Define the struct.
	f.Comment(wrapDocComment(classDocs))
	f.Type().Id(className).Struct(
		jen.Qual(servicePkg, "BaseClass"),
	)

	// Define the constructor.
	constructorName := "New" + className
	f.Comment(fmt.Sprintf("%v creates a new %v.", constructorName, className))
	f.Func().Id(constructorName).Params(
		jen.Id("id").Uint64(),
		jen.Id("client").Op("*").Qual(krpcPkg, "KRPCClient"),
	).Op("*").Id(className).Block(
		jen.Id("c").Op(":=").Op("&").Id(className).Values(jen.Dict{
			jen.Id("BaseClass"): jen.Qual(servicePkg, "BaseClass").Values(jen.Dict{
				jen.Id("Client"): jen.Id("client"),
			}),
		}),
		jen.Id("c").Dot("SetID").Call(jen.Id("id")),
		jen.Return(jen.Id("c")),
	)
	return nil
}
