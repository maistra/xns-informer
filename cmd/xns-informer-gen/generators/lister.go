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
	sw.Do(typedLister, data)
	sw.Do(typedNamespaceLister, data)
	sw.Do(typedHelpers, data)

	return sw.Error()
}

// TODO: Fix resource() package.
var typedInformer = `
type $.type|private$Informer struct {
	factory xnsinformers.InformerFactory
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
$- if .namespaced$
	return &$.type|private$Lister{lister: f.factory.NamespacedResource(f.resource()).Lister()}
$- else$
	return &$.type|private$Lister{lister: f.factory.ClusterResource(f.resource()).Lister()}
$- end$
}
`

var typedLister = `
type $.type|private$Lister struct {
	lister cache.GenericLister
}

var _ listers.$.type|public$Lister = &$.type|private$Lister{}

func (l *$.type|private$Lister) List(selector labels.Selector) (res []*$.version$.$.type|public$, err error) {
	return list$.type|public$(l.lister, selector)
}

$ if .namespaced$
func (l *$.type|private$Lister) $.type|publicPlural$(namespace string) listers.$.type|public$NamespaceLister {
	return &$.type|private$NamespaceLister{lister: l.lister.ByNamespace(namespace)}
}
$- else$
func (l *$.type|private$Lister) Get(name string) (*$.version$.$.type|public$, error) {
	obj, err := l.lister.Get(name)
	if err != nil {
		return nil, err
	}

    out := &$.version$.$.type|public${}
	if err := xnsinformers.ConvertUnstructured(obj, out); err != nil {
        return nil, err
    }

    return out, nil
}
$- end$
`

var typedNamespaceLister = `
$- if .namespaced$
type $.type|private$NamespaceLister struct {
	lister cache.GenericNamespaceLister
}

var _ listers.$.type|public$NamespaceLister = &$.type|private$NamespaceLister{}

func (l *$.type|private$NamespaceLister) List(selector labels.Selector) (res []*$.version$.$.type|public$, err error) {
	return list$.type|public$(l.lister, selector)
}

func (l *$.type|private$NamespaceLister) Get(name string) (*$.version$.$.type|public$, error) {
	obj, err := l.lister.Get(name)
	if err != nil {
		return nil, err
	}

    out := &$.version$.$.type|public${}
	if err := xnsinformers.ConvertUnstructured(obj, out); err != nil {
        return nil, err
    }

    return out, nil
}
$- end$
`

var typedHelpers = `
func list$.type|public$(l xnsinformers.SimpleLister, s labels.Selector) (res []*$.version$.$.type|public$, err error) {
	objects, err := l.List(s)
	if err != nil {
		return nil, err
	}

	for _, obj := range objects {
        out := &$.version$.$.type|public${}
        if err := xnsinformers.ConvertUnstructured(obj, out); err != nil {
            return nil, err
        }

		res = append(res, out)
	}

	return res, nil
}
`
