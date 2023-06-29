/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//nolint:lll
package generators

import (
	"io"

	"k8s.io/code-generator/cmd/client-gen/generators/util"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// versionInterfaceGenerator generates the per-version interface file.
type versionInterfaceGenerator struct {
	generator.DefaultGen
	outputPackage             string
	informersPackage          string
	groupPackage              string
	imports                   namer.ImportTracker
	types                     []*types.Type
	filtered                  bool
	internalInterfacesPackage string
	version                   string
}

var _ generator.Generator = &versionInterfaceGenerator{}

func (g *versionInterfaceGenerator) Filter(c *generator.Context, t *types.Type) bool {
	if !g.filtered {
		g.filtered = true
		return true
	}
	return false
}

func (g *versionInterfaceGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *versionInterfaceGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	return
}

func (g *versionInterfaceGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")

	internalInterfacesPkg := g.informersPackage + "/internalinterfaces"
	versionPkg := g.informersPackage + "/" + g.groupPackage + "/" + g.version
	m := map[string]interface{}{
		"xnsNamespaceSet":                 c.Universe.Type(xnsNamespaceSet),
		"informersInterface":              c.Universe.Type(types.Name{Package: versionPkg, Name: "Interface"}),
		"interfacesTweakListOptionsFunc":  c.Universe.Type(types.Name{Package: internalInterfacesPkg, Name: "TweakListOptionsFunc"}),
		"interfacesSharedInformerFactory": c.Universe.Type(types.Name{Package: internalInterfacesPkg, Name: "SharedInformerFactory"}),
	}

	sw.Do(versionTemplate, m)
	for _, typeDef := range g.types {
		tags, err := util.ParseClientGenTags(append(typeDef.SecondClosestCommentLines, typeDef.CommentLines...))
		if err != nil {
			return err
		}
		m["namespaced"] = !tags.NonNamespaced
		m["type"] = typeDef
		m["versionedType"] = c.Universe.Type(types.Name{Package: versionPkg, Name: typeDef.Name.Name})
		sw.Do(versionFuncTemplate, m)
	}

	return sw.Error()
}

var versionTemplate = `
type version struct {
	factory $.interfacesSharedInformerFactory|raw$
    namespaces $.xnsNamespaceSet|raw$
	tweakListOptions $.interfacesTweakListOptionsFunc|raw$
}

// New returns a new Interface.
func New(f $.interfacesSharedInformerFactory|raw$, namespaces $.xnsNamespaceSet|raw$, tweakListOptions $.interfacesTweakListOptionsFunc|raw$) $.informersInterface|raw$ {
	return &version{factory: f, namespaces: namespaces, tweakListOptions: tweakListOptions}
}
`

var versionFuncTemplate = `
// $.type|publicPlural$ returns a $.type|public$Informer.
func (v *version) $.type|publicPlural$() $.versionedType|raw$Informer {
	return &$.type|private$Informer{factory: v.factory$if .namespaced$, namespaces: v.namespaces$end$, tweakListOptions: v.tweakListOptions}
}
`
