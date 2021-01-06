package generators

import (
	"io"
	"sort"
	"strings"

	clientgentypes "k8s.io/code-generator/cmd/client-gen/types"
	codegennamer "k8s.io/code-generator/pkg/namer"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// genericGenerator generates the generic informer.
type genericGenerator struct {
	generator.DefaultGen
	outputPackage        string
	imports              namer.ImportTracker
	groupVersions        map[string]clientgentypes.GroupVersions
	groupGoNames         map[string]string
	pluralExceptions     map[string]string
	typesForGroupVersion map[clientgentypes.GroupVersion][]*types.Type
	filtered             bool
}

var _ generator.Generator = &genericGenerator{}

func (g *genericGenerator) Filter(c *generator.Context, t *types.Type) bool {
	if !g.filtered {
		g.filtered = true
		return true
	}
	return false
}

func (g *genericGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw":                namer.NewRawNamer(g.outputPackage, g.imports),
		"allLowercasePlural": namer.NewAllLowercasePluralNamer(g.pluralExceptions),
		"publicPlural":       namer.NewPublicPluralNamer(g.pluralExceptions),
		"resource":           codegennamer.NewTagOverrideNamer("resourceName", namer.NewAllLowercasePluralNamer(g.pluralExceptions)),
	}
}

func (g *genericGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	imports = append(imports, "fmt")
	return
}

type group struct {
	GroupGoName string
	Name        string
	Versions    []*version
}

type groupSort []group

func (g groupSort) Len() int { return len(g) }
func (g groupSort) Less(i, j int) bool {
	return strings.ToLower(g[i].Name) < strings.ToLower(g[j].Name)
}
func (g groupSort) Swap(i, j int) { g[i], g[j] = g[j], g[i] }

type version struct {
	Name      string
	GoName    string
	Resources []*types.Type
}

type versionSort []*version

func (v versionSort) Len() int { return len(v) }
func (v versionSort) Less(i, j int) bool {
	return strings.ToLower(v[i].Name) < strings.ToLower(v[j].Name)
}
func (v versionSort) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

func (g *genericGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "{{", "}}")

	groups := []group{}
	schemeGVs := make(map[*version]*types.Type)

	orderer := namer.Orderer{Namer: namer.NewPrivateNamer(0)}
	for groupPackageName, groupVersions := range g.groupVersions {
		group := group{
			GroupGoName: g.groupGoNames[groupPackageName],
			Name:        groupVersions.Group.NonEmpty(),
			Versions:    []*version{},
		}
		for _, v := range groupVersions.Versions {
			gv := clientgentypes.GroupVersion{Group: groupVersions.Group, Version: v.Version}
			version := &version{
				Name:      v.Version.NonEmpty(),
				GoName:    namer.IC(v.Version.NonEmpty()),
				Resources: orderer.OrderTypes(g.typesForGroupVersion[gv]),
			}
			func() {
				schemeGVs[version] = c.Universe.Variable(types.Name{Package: g.typesForGroupVersion[gv][0].Name.Package, Name: "SchemeGroupVersion"})
			}()
			group.Versions = append(group.Versions, version)
		}
		sort.Sort(versionSort(group.Versions))
		groups = append(groups, group)
	}
	sort.Sort(groupSort(groups))

	m := map[string]interface{}{
		"cacheGenericLister":         c.Universe.Type(types.Name{Package: "k8s.io/client-go/tools/cache", Name: "GenericLister"}),
		"cacheNewGenericLister":      c.Universe.Function(types.Name{Package: "k8s.io/client-go/tools/cache", Name: "NewGenericLister"}),
		"cacheSharedIndexInformer":   c.Universe.Type(types.Name{Package: "k8s.io/client-go/tools/cache", Name: "SharedIndexInformer"}),
		"groups":                     groups,
		"schemeGVs":                  schemeGVs,
		"schemaGroupResource":        c.Universe.Type(types.Name{Package: "k8s.io/apimachinery/pkg/runtime/schema", Name: "GroupResource"}),
		"schemaGroupVersionResource": c.Universe.Type(types.Name{Package: "k8s.io/apimachinery/pkg/runtime/schema", Name: "GroupVersionResource"}),
	}

	sw.Do(genericInformer, m)
	sw.Do(forResource, m)

	return sw.Error()
}

var genericInformer = `
// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() {{.cacheSharedIndexInformer|raw}}
	Lister() {{.cacheGenericLister|raw}}
}
type genericInformer struct {
	informer {{.cacheSharedIndexInformer|raw}}
	resource {{.schemaGroupResource|raw}}
}
// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() {{.cacheSharedIndexInformer|raw}} {
	return f.informer
}
// Lister returns the GenericLister.
func (f *genericInformer) Lister() {{.cacheGenericLister|raw}} {
	return {{.cacheNewGenericLister|raw}}(f.Informer().GetIndexer(), f.resource)
}
`

var forResource = `
// ForResource gives generic access to a shared informer of the matching type
func (f *sharedInformerFactory) ForResource(resource {{.schemaGroupVersionResource|raw}}) (GenericInformer, error) {
	switch resource {
		{{range $group := .groups -}}{{$GroupGoName := .GroupGoName -}}
			{{range $version := .Versions -}}
	// Group={{$group.Name}}, Version={{.Name}}
				{{range .Resources -}}
	case {{index $.schemeGVs $version|raw}}.WithResource("{{.|resource}}"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.{{$GroupGoName}}().{{$version.GoName}}().{{.|publicPlural}}().Informer()}, nil
				{{end}}
			{{end}}
		{{end -}}
	}
	return nil, fmt.Errorf("no informer found for %v", resource)
}
`
