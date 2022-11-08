package gen

import (
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
