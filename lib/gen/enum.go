package gen

import (
	"fmt"

	"github.com/atburke/krpc-go/lib/api"
	"github.com/atburke/krpc-go/lib/utils"
	"github.com/dave/jennifer/jen"
	"github.com/mitchellh/go-wordwrap"
	"github.com/ztrue/tracerr"
)

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
			jen.Id(valueName).Id(enumName).Op("=").Lit(int(value.Value)),
		)
	}

	f.Const().Defs(defs...)
	return nil
}
