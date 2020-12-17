package generators

import (
	"fmt"
	"io"
	"path/filepath"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

type factoryGenerator struct {
	generator.DefaultGen
	outputPackage string
	groups        []string
	imports       namer.ImportTracker
}

type groupData struct {
	Name      string
	Interface *types.Type
	New       *types.Type
}

func (g *factoryGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *factoryGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	imports = append(imports, "k8s.io/apimachinery/pkg/runtime/schema")
	imports = append(imports, "k8s.io/client-go/tools/cache")
	imports = append(imports, fmt.Sprintf("xnsinformers %q", xnsinformersPkg))
	return
}

func (g *factoryGenerator) Init(c *generator.Context, w io.Writer) error {
	groups := []groupData{}

	for _, group := range g.groups {
		if group == "" {
			continue
		}

		groupPackage := filepath.Join(g.outputPackage, group)
		iface := c.Universe.Type(types.Name{Package: groupPackage, Name: "Interface"})

		groups = append(groups, groupData{
			Name:      namer.IC(group),
			Interface: iface,
			New:       c.Universe.Function(types.Name{Package: groupPackage, Name: "New"}),
		})
	}

	data := map[string]interface{}{
		"groups": groups,
	}

	sw := generator.NewSnippetWriter(w, c, "$", "$")
	sw.Do(factoryInterface, data)
	return sw.Error()
}

var factoryInterface = `
type SharedInformerFactory interface {
$- range .groups$
  $.Name$() $.Interface|raw$
$- end$
  ForResource(resource schema.GroupVersionResource) (GenericInformer, error)
}

type sharedInformerFactory struct {
  factory xnsinformers.SharedInformerFactory
}

// NewSharedInformerFactory returns a new SharedInformerFactory.
func NewSharedInformerFactory(f xnsinformers.SharedInformerFactory) SharedInformerFactory {
  return &sharedInformerFactory{factory: f}
}

$- range .groups$
// $.Name$ returns a new $.Interface|raw$.
func (f *sharedInformerFactory) $.Name$() $.Interface|raw$ {
	return $.New|raw$(f.factory)
}
$- end$
`
