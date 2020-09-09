package generators

import (
	"fmt"
	"io"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/types"
)

type versionInterfaceGenerator struct {
	generator.DefaultGen
	informersPackage string
	groupVersion     GroupVersion
	types            []*types.Type
}

func (g *versionInterfaceGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, fmt.Sprintf("informers %q", g.informersPackage))
	imports = append(imports, fmt.Sprintf("xnsinformers %q", xnsinformersPkg))
	return
}

func (g *versionInterfaceGenerator) Init(c *generator.Context, w io.Writer) error {
	data := map[string]interface{}{
		"types":   g.types,
		"group":   g.groupVersion.Group,
		"version": g.groupVersion.Version,
	}

	sw := generator.NewSnippetWriter(w, c, "$", "$")
	sw.Do(versionInterface, data)
	return sw.Error()
}

var versionInterface = `
type Interface interface {
$- range .types$
    $.|publicPlural$() informers.$.|public$Informer
$- end$
}

type version struct {
    factory xnsinformers.InformerFactory
}

func New(factory xnsinformers.InformerFactory) Interface {
  return &version{factory: factory}
}

$- range .types$
func (v *version) $.|publicPlural$() informers.$.|public$Informer {
    return &$.|private$Informer{factory: v.factory}
}
$- end$
`
