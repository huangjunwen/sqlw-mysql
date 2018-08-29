package render

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/beevik/etree"

	"github.com/huangjunwen/sqlw-mysql/datasrc"
	"github.com/huangjunwen/sqlw-mysql/infos"
)

// Renderer is used to render templates to final source code.
type Renderer struct {
	// Options.
	dsn       string
	stmtDir   string
	tmplDir   http.FileSystem
	outputDir string
	outputPkg string
	whitelist map[string]struct{}
	blacklist map[string]struct{}

	// Runtime vars.
	loader   *datasrc.Loader
	db       *infos.DBInfo
	manifest *Manifest
}

// Option is used to create Renderer.
type Option func(*Renderer) error

// DSN sets the data source name. (required)
func DSN(dsn string) Option {
	return func(r *Renderer) error {
		r.dsn = dsn
		return nil
	}
}

// StmtDir sets the statement xml directory.
func StmtDir(stmtDir string) Option {
	return func(r *Renderer) error {
		stmtDir = path.Clean(stmtDir)
		fi, err := os.Stat(stmtDir)
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			return fmt.Errorf("%+q is not a directory.", stmtDir)
		}
		r.stmtDir = stmtDir
		return nil
	}
}

// TmplDir sets the template directory. (required)
func TmplDir(tmplDir http.FileSystem) Option {
	return func(r *Renderer) error {
		r.tmplDir = tmplDir
		return nil
	}
}

// OutputDir sets the output directory. It will mkdir if not exists. (required)
func OutputDir(outputDir string) Option {
	return func(r *Renderer) error {
		if outputDir == "" {
			return fmt.Errorf("Output directory is empty.")
		}
		outputDir = path.Clean(outputDir)
		if outputDir == "/" {
			return fmt.Errorf("Output directory can't be root '/'")
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}
		r.outputDir = outputDir
		return nil
	}
}

// OutputPkg sets an alternative package name for the generated code. By default the output directory is used.
func OutputPkg(outputPkg string) Option {
	return func(r *Renderer) error {
		r.outputPkg = outputPkg
		return nil
	}
}

// Whitelist sets the whitelist of table names to render.
func Whitelist(whitelist []string) Option {
	return func(r *Renderer) error {
		if len(r.blacklist) != 0 {
			return fmt.Errorf("Can't set whitelist since blacklist is set")
		}
		r.whitelist = make(map[string]struct{})
		for _, item := range whitelist {
			r.whitelist[item] = struct{}{}
		}
		return nil
	}

}

// Blacklist sets the blacklist of table names to render.
func Blacklist(blacklist []string) Option {
	return func(r *Renderer) error {
		if len(r.whitelist) != 0 {
			return fmt.Errorf("Can't set blacklist since whitelist is set")
		}
		r.blacklist = make(map[string]struct{})
		for _, item := range blacklist {
			r.blacklist[item] = struct{}{}
		}
		return nil
	}
}

// NewRenderer creates a new Renderer.
func NewRenderer(opts ...Option) (*Renderer, error) {

	r := &Renderer{}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	if r.dsn == "" {
		return nil, fmt.Errorf("Missing DSN")
	}
	if r.tmplDir == nil {
		return nil, fmt.Errorf("Missing TmplDir")
	}
	if r.outputDir == "" {
		return nil, fmt.Errorf("Missing OutputDir")
	}
	if r.outputPkg == "" {
		r.outputPkg = path.Base(r.outputDir)
	}
	return r, nil
}

// Run the render process.
func (r *Renderer) Run() error {

	r.loader = nil
	r.db = nil
	r.manifest = nil

	// Create loader.
	var err error
	r.loader, err = datasrc.NewLoader(r.dsn)
	if err != nil {
		return err
	}
	defer r.loader.Close()

	// Load db.
	r.db, err = infos.NewDBInfo(r.loader)
	if err != nil {
		return err
	}

	// Load manifest.
	r.manifest, err = LoadManifest(r.tmplDir, r.funcMap())
	if err != nil {
		return err
	}

	// Render per run templates.
	for _, tmpls := range r.manifest.PerRun {
		if err := r.render(tmpls, map[string]interface{}{
			"PackageName": r.outputPkg,
			"DB":          r.db,
		}); err != nil {
			return err
		}
	}

	// Render per table templates.
	for _, table := range r.db.Tables() {

		// Filter table.
		if len(r.whitelist) != 0 {
			if _, found := r.whitelist[table.TableName()]; !found {
				continue
			}
		} else if len(r.blacklist) != 0 {
			if _, found := r.blacklist[table.TableName()]; found {
				continue
			}
		}

		// Render.
		for _, tmpls := range r.manifest.PerTable {
			if err := r.render(tmpls, map[string]interface{}{
				"PackageName": r.outputPkg,
				"DB":          r.db,
				"Table":       table,
			}); err != nil {
				return err
			}
		}
	}

	if r.stmtDir == "" {
		return nil
	}

	// Render per stmt xml templates.
	stmtFileInfos, err := ioutil.ReadDir(r.stmtDir)
	if err != nil {
		return err
	}

	for _, stmtFileInfo := range stmtFileInfos {

		// Skip directory and non xml files.
		if stmtFileInfo.IsDir() {
			continue
		}
		stmtFileName := stmtFileInfo.Name()
		if !strings.HasSuffix(stmtFileName, ".xml") {
			continue
		}

		// Load xml.
		doc := etree.NewDocument()
		if err := doc.ReadFromFile(path.Join(r.stmtDir, stmtFileName)); err != nil {
			return err
		}

		// Xml -> StmtInfo.
		stmtInfos := []*infos.StmtInfo{}
		for _, elem := range doc.ChildElements() {
			stmtInfo, err := infos.NewStmtInfo(r.loader, r.db, elem)
			if err != nil {
				return err
			}
			stmtInfos = append(stmtInfos, stmtInfo)
		}

		// Render.
		for _, tmpls := range r.manifest.PerStmtXML {
			if err := r.render(tmpls, map[string]interface{}{
				"PackageName": r.outputPkg,
				"DB":          r.db,
				"Stmts":       stmtInfos,
				"StmtXMLName": strings.TrimSuffix(stmtFileName, ".xml"),
			}); err != nil {
				return err
			}
		}

	}

	return nil

}

func (r *Renderer) render(tmpls [2]*template.Template, data interface{}) error {

	nameTmpl, contentTmpl := tmpls[0], tmpls[1]
	nameBuf := bytes.Buffer{}
	contentBuf := bytes.Buffer{}

	// Render file name.
	if err := nameTmpl.Execute(&nameBuf, data); err != nil {
		return err
	}

	// Render file content.
	if err := contentTmpl.Execute(&contentBuf, data); err != nil {
		return err
	}

	// Open file to write.
	f, err := os.OpenFile(path.Join(r.outputDir, string(nameBuf.Bytes())), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write.
	_, err = f.Write(contentBuf.Bytes())
	if err != nil {
		return err
	}

	return nil

}
