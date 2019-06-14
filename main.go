//go:generate esc -o templates.go templates
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime/debug"
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
	version   bool
)

func main() {
	// Parse flags.
	flag.StringVar(&dsn, "dsn", "", "(Required) Data source name. e.g. \"user:passwd@tcp(host:port)/db?parseTime=true\"")
	flag.StringVar(&tmplDir, "tmpl", "", "(Optional) Custom templates directory. Or use '@name' to use the named builtin template.")
	flag.StringVar(&stmtDir, "stmt", "", "(Optional) Statement xmls directory.")
	flag.StringVar(&outputDir, "out", "models", "(Optional) Output directory for generated code.")
	flag.StringVar(&outputPkg, "pkg", "", "(Optional) Alternative package name of the generated code.")
	flag.Var(&whitelist, "whitelist", "(Optional) Comma separated table names to render.")
	flag.Var(&blacklist, "blacklist", "(Optional) Comma separated table names not to render.")
	flag.BoolVar(&version, "version", false, "Show version information.")
	flag.Parse()

	if version {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Println("sqlw-mysql unknown version")
		} else {
			fmt.Printf("sqlw-mysql version=%s sum=%s\n", info.Main.Version, info.Main.Sum)
		}
		return
	}

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
