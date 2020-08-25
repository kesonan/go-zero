package gogen

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tal-tech/go-zero/core/collection"
	"github.com/tal-tech/go-zero/tools/goctl/rpc/parser"
	"github.com/tal-tech/go-zero/tools/goctl/util/templatex"
)

var (
	logicTemplate = `package logic

import (
	"context"

	{{.imports}}
	"github.com/tal-tech/go-zero/core/logx"
)

type (
	{{.logicName}} struct {
		ctx context.Context
		logx.Logger
		// todo: add your logic here and delete this line
	}
)

func New{{.logicName}}(ctx context.Context,svcCtx *svc.ServiceContext) *{{.logicName}} {
	return &{{.logicName}}{
		ctx:    ctx,
		Logger: logx.WithContext(ctx),
		// todo: add your logic here and delete this line
	}
}
{{.functions}}
`
	logicFunctionTemplate = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} (in *{{.package}}.{{.request}}) (*{{.package}}.{{.response}}, error) {
	var resp {{.package}}.{{.response}}
	// todo: add your logic here and delete this line
	
	return &resp,nil
}
`
)

func (g *defaultRpcGenerator) genLogic() error {
	logicPath := g.dirM[dirLogic]
	protoPkg := g.ast.Package
	service := g.ast.Service
	for _, item := range service {
		for _, method := range item.Funcs {
			logicName := fmt.Sprintf("%slogic.go", method.Name.Lower())
			filename := filepath.Join(logicPath, logicName)
			functions, err := genLogicFunction(protoPkg, method)
			if err != nil {
				return err
			}
			imports := collection.NewSet()
			pbImport := fmt.Sprintf(`%v "%v"`, protoPkg, g.mustGetPackage(dirPb))
			svcImport := fmt.Sprintf(`"%v"`, g.mustGetPackage(dirSvc))
			imports.AddStr(pbImport, svcImport)
			err = templatex.With("logic").GoFmt(true).Parse(logicTemplate).SaveTo(map[string]interface{}{
				"logicName": fmt.Sprintf("%sLogic", method.Name),
				"functions": functions,
				"imports":   strings.Join(imports.KeysStr(), "\r\n"),
			}, filename, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func genLogicFunction(packageName string, method *parser.Func) (string, error) {
	var functions = make([]string, 0)
	buffer, err := templatex.With("fun").Parse(logicFunctionTemplate).Execute(map[string]interface{}{
		"logicName":  fmt.Sprintf("%sLogic", method.Name),
		"method":     method.Name,
		"package":    packageName,
		"request":    method.InType,
		"response":   method.OutType,
		"hasComment": len(method.Document) > 0,
		"comment":    strings.Join(method.Document, "\r\n"),
	})
	if err != nil {
		return "", err
	}
	functions = append(functions, buffer.String())
	return strings.Join(functions, "\n"), nil
}