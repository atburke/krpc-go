package gen

import (
	"fmt"

	"github.com/atburke/krpc-go/lib/utils"
	"github.com/atburke/krpc-go/types"
	"github.com/dave/jennifer/jen"
	"github.com/mitchellh/go-wordwrap"
	"github.com/ztrue/tracerr"
)

// GenerateEnum generates an enum for a given enum definition.
func GenerateEnum(f *jen.File, enum *types.Enumeration) error {
	enumName := enum.Name
	enumDocs, err := utils.ParseXMLDocumentation(enum.Documentation, enumName+" - ")
	if err != nil {
		return tracerr.Wrap(err)
	}

	// Define the enum type
	f.Comment(wordwrap.WrapString(enumDocs, DocsLineLength))
	f.Type().Id(enumName).Int32()

	// Define the enum values
	var defs []jen.Code
	for _, value := range enum.Values {
		valueName := fmt.Sprintf("%v_%v", enumName, value.Name)
		valueDocs, err := utils.ParseXMLDocumentation(value.Documentation, "")
		if err != nil {
			return tracerr.Wrap(err)
		}
		defs = append(defs,
			jen.Comment(wrapDocComment(valueDocs)),
			jen.Id(valueName).Id(enumName).Op("=").Lit(int(value.Value)),
		)
	}

	f.Const().Defs(defs...)

	// Fill out enum interface
	f.Func().Params(jen.Id("v").Id(enumName)).Id("Value").Params().Int32().Block(
		jen.Return(jen.Int32().Call(jen.Id("v"))),
	)
	f.Func().Params(jen.Id("v").Op("*").Id(enumName)).Id("SetValue").Params(jen.Id("val").Int32()).Block(
		jen.Op("*").Id("v").Op("=").Id(enumName).Call(jen.Id("val")),
	)
	return nil
}
