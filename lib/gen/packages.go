package gen

import "strings"

const (
	apiPkg     = "github.com/atburke/krpc-go/lib/api"
	krpcPkg    = "github.com/atburke/krpc-go"
	servicePkg = "github.com/atburke/krpc-go/lib/service"
	encodePkg  = "github.com/atburke/krpc-go/lib/encode"
	tracerrPkg = "github.com/ztrue/tracerr"
)

func getServicePackage(serviceName string) string {
	return "github.com/atburke/krpc-go/" + strings.ToLower(serviceName)
}
