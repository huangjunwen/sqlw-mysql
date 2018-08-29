package render

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
)

// Manifest contains file and template information.
type Manifest struct {
	dir     http.FileSystem
	funcMap template.FuncMap

	JSON struct {
		ScanTypeMap string   `json:"scanTypeMap"`
		PerRun      []string `json:"perRun"`
		PerTable    []string `json:"perTable"`
		PerStmtXML  []string `json:"perStmtXML"`
	}

	ScanTypeMap ScanTypeMap
	PerRun      [][2]*template.Template // [nameTmpl, contentTmpl]
	PerTable    [][2]*template.Template
	PerStmtXML  [][2]*template.Template
}

// LoadManifest loads file and templates from a manifest in a directory.
func LoadManifest(dir http.FileSystem, funcMap template.FuncMap) (*Manifest, error) {

	manif := &Manifest{
		dir:     dir,
		funcMap: funcMap,
	}

	// Load manifest.json.
	if err := manif.loadManifest(); err != nil {
		return nil, err
	}

	// Load ScanTypeMap.
	if err := manif.loadScanTypeMap(); err != nil {
		return nil, err
	}

	// Load perRun templates.
	for _, name := range manif.JSON.PerRun {
		if err := manif.loadPerRunTmpl(name); err != nil {
			return nil, err
		}
	}

	// Load perTable templates.
	for _, name := range manif.JSON.PerTable {
		if err := manif.loadPerTableTmpl(name); err != nil {
			return nil, err
		}
	}

	// Load perStmtXML templates.
	for _, name := range manif.JSON.PerStmtXML {
		if err := manif.loadPerStmtXMLTmpl(name); err != nil {
			return nil, err
		}
	}

	return manif, nil

}

func (manif *Manifest) loadManifest() error {

	f, err := manif.dir.Open("manifest.json")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&manif.JSON); err != nil {
		return err
	}

	return nil
}

func (manif *Manifest) loadScanTypeMap() error {

	if manif.JSON.ScanTypeMap == "" {
		return nil
	}

	f, err := manif.dir.Open(manif.JSON.ScanTypeMap)
	if err != nil {
		return err
	}
	defer f.Close()

	scanTypeMap, err := LoadScanTypeMap(f)
	if err != nil {
		return err
	}
	manif.ScanTypeMap = scanTypeMap

	return nil
}

func (manif *Manifest) loadPerRunTmpl(name string) error {
	nameTmpl, contentTmpl, err := manif.loadTmpl(name)
	if err != nil {
		return err
	}
	manif.PerRun = append(manif.PerRun, [2]*template.Template{nameTmpl, contentTmpl})
	return nil
}

func (manif *Manifest) loadPerTableTmpl(name string) error {
	nameTmpl, contentTmpl, err := manif.loadTmpl(name)
	if err != nil {
		return err
	}
	manif.PerTable = append(manif.PerTable, [2]*template.Template{nameTmpl, contentTmpl})
	return nil
}

func (manif *Manifest) loadPerStmtXMLTmpl(name string) error {
	nameTmpl, contentTmpl, err := manif.loadTmpl(name)
	if err != nil {
		return err
	}
	manif.PerStmtXML = append(manif.PerStmtXML, [2]*template.Template{nameTmpl, contentTmpl})
	return nil
}

func (manif *Manifest) loadTmpl(name string) (*template.Template, *template.Template, error) {

	f, err := manif.dir.Open(name)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}

	nameTmpl, err := template.New(name + "@name").Funcs(manif.funcMap).Parse(strings.TrimSuffix(name, ".tmpl"))
	if err != nil {
		return nil, nil, err
	}

	contentTmpl, err := template.New(name + "@content").Funcs(manif.funcMap).Parse(string(content))
	if err != nil {
		return nil, nil, err
	}

	return nameTmpl, contentTmpl, nil

}
