package generators

import (
	"fmt"
	"io"
	"path/filepath"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

type groupInterfaceGenerator struct {
	generator.DefaultGen
	outputPackage string
	group         string
	versions      []string
	imports       namer.ImportTracker
}

type versionData struct {
	Name      string
	Interface *types.Type
	New       *types.Type
}

func (g *groupInterfaceGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *groupInterfaceGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	imports = append(imports, fmt.Sprintf("xnsinformers %q", xnsinformersPkg))
	return
}

func (g *groupInterfaceGenerator) Init(c *generator.Context, w io.Writer) error {
	versions := []versionData{}

	for _, v := range g.versions {
		if v == "" {
			continue
		}

		versionPackage := filepath.Join(g.outputPackage, v)
		iface := c.Universe.Type(types.Name{Package: versionPackage, Name: "Interface"})

		versions = append(versions, versionData{
			Name:      namer.IC(v),
			Interface: iface,
			New:       c.Universe.Function(types.Name{Package: versionPackage, Name: "New"}),
		})
	}

	data := map[string]interface{}{
		"group":    g.group,
		"versions": versions,
	}

	sw := generator.NewSnippetWriter(w, c, "$", "$")
	sw.Do(groupInterface, data)
	return sw.Error()
}

var groupInterface = `
type Interface interface {
$- range .versions$
  $.Name$() $.Interface|raw$
$- end$
}

type group struct {
  factory xnsinformers.InformerFactory
}

func New(factory xnsinformers.InformerFactory) Interface {
  return &group{factory: factory}
}

$- range .versions$
// $.Name$ returns a new $.Interface|raw$.
func (g *group) $.Name$() $.Interface|raw$ {
	return $.New|raw$(g.factory)
}
$- end$
`
