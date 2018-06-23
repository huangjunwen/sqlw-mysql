package render

import (
	"bytes"
	"fmt"
	"go/format"
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
	dsn       string
	tmplDir   http.FileSystem
	stmtDir   string
	outputDir string
	outputPkg string
	whitelist map[string]struct{}
	blacklist map[string]struct{}

	loader      *datasrc.Loader
	db          *infos.DBInfo
	manifest    *Manifest
	scanTypeMap ScanTypeMap
	headnote    string
	templates   map[string]*template.Template // name -> template
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

// TmplDir sets the template directory. (required)
func TmplDir(tmplDir http.FileSystem) Option {
	return func(r *Renderer) error {
		r.tmplDir = tmplDir
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
	r.scanTypeMap = nil
	r.headnote = ""
	r.templates = make(map[string]*template.Template)

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
	manifestFile, err := r.tmplDir.Open("manifest.json")
	if err != nil {
		return err
	}
	defer manifestFile.Close()

	r.manifest, err = NewManifest(manifestFile)
	if err != nil {
		return err
	}

	// Load scanTypeMap.
	scanTypeMapFile, err := r.tmplDir.Open(r.manifest.ScanTypeMap)
	if err != nil {
		return err
	}

	r.scanTypeMap, err = NewScanTypeMap(scanTypeMapFile)
	if err != nil {
		return err
	}

	// Load headnote.
	if r.manifest.Headnote != "" {
		headnoteFile, err := r.tmplDir.Open(r.manifest.Headnote)
		if err != nil {
			return err
		}
		defer headnoteFile.Close()

		headnoteContent, err := ioutil.ReadAll(headnoteFile)
		if err != nil {
			return err
		}

		r.headnote = string(headnoteContent)
		if r.headnote != "" {
			r.headnote += "\n" // Add a newline.
		}
	}

	// Start renderring.

	// Render tables.
	for _, table := range r.db.Tables() {

		// Filter.
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
		if err := r.render(
			r.manifest.Templates.Table,
			fmt.Sprintf("table_%s.go", table.TableName()),
			map[string]interface{}{
				"PackageName": r.outputPkg,
				"DB":          r.db,
				"Table":       table,
			}); err != nil {
			return err
		}

		if r.manifest.Templates.TableTest == "" {
			continue
		}

		// Render test.
		if err := r.render(
			r.manifest.Templates.TableTest,
			fmt.Sprintf("table_%s_test.go", table.TableName()),
			map[string]interface{}{
				"PackageName": r.outputPkg,
				"DB":          r.db,
				"Table":       table,
			}); err != nil {
			return err
		}

	}

	// Render statements.
	if r.stmtDir != "" {

		stmtFileInfos, err := ioutil.ReadDir(r.stmtDir)
		if err != nil {
			return err
		}
		for _, stmtFileInfo := range stmtFileInfos {

			// Skip directory and non-xml files.
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
			if err := r.render(
				r.manifest.Templates.Stmt,
				fmt.Sprintf("stmt_%s.go", stripSuffix(stmtFileName)),
				map[string]interface{}{
					"PackageName": r.outputPkg,
					"DB":          r.db,
					"Stmts":       stmtInfos,
				}); err != nil {
				return err
			}

			if r.manifest.Templates.StmtTest == "" {
				continue
			}

			// Render test.
			if err := r.render(
				r.manifest.Templates.StmtTest,
				fmt.Sprintf("stmt_%s_test.go", stripSuffix(stmtFileName)),
				map[string]interface{}{
					"PackageName": r.outputPkg,
					"DB":          r.db,
					"Stmts":       stmtInfos,
				}); err != nil {
				return err
			}

		}

	}

	// Render etc files.
	for _, tmplName := range r.manifest.Templates.Etc {

		if err := r.render(
			tmplName,
			fmt.Sprintf("etc_%s.go", stripSuffix(tmplName)),
			map[string]interface{}{
				"PackageName": r.outputPkg,
				"DB":          r.db,
			}); err != nil {
			return err
		}

	}

	return nil
}

func (r *Renderer) render(tmplName, fileName string, data interface{}) error {

	tmpl := r.templates[tmplName]

	// Not exists yet, load the template.
	if tmpl == nil {
		tmplFile, err := r.tmplDir.Open(tmplName)
		if err != nil {
			return err
		}
		defer tmplFile.Close()

		tmplContent, err := ioutil.ReadAll(tmplFile)
		if err != nil {
			return err
		}

		tmpl, err = template.New(tmplName).Funcs(r.funcMap()).Parse(string(tmplContent))
		if err != nil {
			return err
		}

		r.templates[tmplName] = tmpl

	}

	// Open output file.
	file, err := os.OpenFile(path.Join(r.outputDir, fileName), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Output headnote.
	_, err = file.WriteString(r.headnote)
	if err != nil {
		return err
	}

	// Render.
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}

	// Format.
	fmtBuf, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	// Output.
	_, err = file.Write(fmtBuf)
	if err != nil {
		return err
	}

	return nil

}

func stripSuffix(s string) string {
	i := strings.LastIndexByte(s, '.')
	if i < 0 {
		return s
	}
	return s[:i]
}
