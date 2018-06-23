//go:generate esc -o templates.go templates
package main

import (
	"flag"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/huangjunwen/sqlw-mysql/render"
)

type commaSeperatd []string

func (cs *commaSeperatd) String() string {
	return strings.Join(*cs, ",")
}

func (cs *commaSeperatd) Set(s string) error {
	*cs = strings.Split(s, ",")
	return nil
}

var (
	dsn       string
	tmplDir   string
	stmtDir   string
	outputDir string
	outputPkg string
	whitelist commaSeperatd
	blacklist commaSeperatd
)

func main() {
	// Parse flags.
	flag.StringVar(&dsn, "dsn", "", "(Required) Data source name. e.g. \"user:passwd@tcp(host:port)/db?parseTime=true\"")
	flag.StringVar(&tmplDir, "tmpl", "", "(Optional) Custom templates directory.")
	flag.StringVar(&stmtDir, "stmt", "", "(Optional) Statement xmls directory.")
	flag.StringVar(&outputDir, "out", "models", "(Optional) Output directory for generated code.")
	flag.StringVar(&outputPkg, "pkg", "", "(Optional) Alternative package name of the generated code.")
	flag.Var(&whitelist, "whitelist", "(Optional) Comma seperated table names to render.")
	flag.Var(&blacklist, "blacklist", "(Optional) Comma seperated table names not to render.")
	flag.Parse()
	if dsn == "" {
		log.Fatalf("Missing -dsn")
	}

	// Choose template.
	fs := http.FileSystem(nil)
	if tmplDir == "" {
		// Use default builtin template.
		fs = newPrefixFS("/templates/default", FS(false))
	} else if tmplDir[0] == '@' {
		// Use other builtin template.
		fs = newPrefixFS(path.Join("/templates", tmplDir[1:]), FS(false))
	} else {
		// Use custom template.
		fs = http.Dir(tmplDir)
	}

	// Create Renderer.
	renderer, err := render.NewRenderer(
		render.DSN(dsn),
		render.TmplDir(fs),
		render.StmtDir(stmtDir),
		render.OutputDir(outputDir),
		render.OutputPkg(outputPkg),
		render.Whitelist([]string(whitelist)),
		render.Blacklist([]string(blacklist)),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Run!
	if err := renderer.Run(); err != nil {
		log.Fatal(err)
	}

	return
}
