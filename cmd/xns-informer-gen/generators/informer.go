package generators

import (
	"fmt"
	"io"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"

	"k8s.io/code-generator/cmd/client-gen/generators/util"
)

type listerGenerator struct {
	generator.DefaultGen
	imports          namer.ImportTracker
	listersPackage   string
	informersPackage string
	groupVersion     GroupVersion
	typeToGenerate   *types.Type
}

func (g *listerGenerator) Filter(c *generator.Context, t *types.Type) bool {
	return t == g.typeToGenerate
}

func (g *listerGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	imports = append(imports, fmt.Sprintf("listers %q", g.listersPackage))
	imports = append(imports, fmt.Sprintf("informers %q", g.informersPackage))
	imports = append(imports, fmt.Sprintf("xnsinformers %q", xnsinformersPkg))
	return
}

func (g *listerGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	tags, err := util.ParseClientGenTags(append(t.SecondClosestCommentLines, t.CommentLines...))
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"type":       t,
		"group":      g.groupVersion.Group,
		"version":    g.groupVersion.Version,
		"namespaced": !tags.NonNamespaced,
	}

	sw := generator.NewSnippetWriter(w, c, "$", "$")
	sw.Do(typedInformer, data)

	return sw.Error()
}

var typedInformer = `
type $.type|private$Informer struct {
	factory xnsinformers.SharedInformerFactory
}

var _ informers.$.type|public$Informer = &$.type|private$Informer{}

func (f *$.type|private$Informer) resource() schema.GroupVersionResource {
	return $.version$.SchemeGroupVersion.WithResource("$.type|allLowercasePlural$")
}

func (f *$.type|private$Informer) Informer() cache.SharedIndexInformer {
$- if .namespaced$
	return f.factory.NamespacedResource(f.resource()).Informer()
$- else$
    return f.factory.ClusterResource(f.resource()).Informer()
$- end$
}

func (f *$.type|private$Informer) Lister() listers.$.type|public$Lister {
    idx := xnsinformers.NewCacheConverter(f.Informer().GetIndexer(), &$.version$.$.type|public${})
    return listers.New$.type|public$Lister(idx)
}
`
