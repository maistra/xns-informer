package testing

import (
	"sort"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

// SortRuntimeObjects sorts the given objects by their namespace and name.
func SortRuntimeObjects(objects []runtime.Object) {
	sort.Slice(objects, func(i, j int) bool {
		iObj, jObj := objects[i], objects[j]

		iKey, err := cache.MetaNamespaceKeyFunc(iObj)
		if err != nil {
			panic(err)
		}

		jKey, err := cache.MetaNamespaceKeyFunc(jObj)
		if err != nil {
			panic(err)
		}

		return iKey < jKey
	})
}

// ObjectMap represents a map of namespaces to a list of objects.
type ObjectMap map[string][]runtime.Object

// Namespaces returns all the namespaces in the map.
func (om ObjectMap) Namespaces() (res []string) {
	for namespace := range om {
		res = append(res, namespace)
	}

	return res
}

// Objects returns all the objects in the map.
func (om ObjectMap) Objects() (res []runtime.Object) {
	for _, objects := range om {
		res = append(res, objects...)
	}

	return res
}

// KeysForNamespace returns the namespace/name pairs for each object in the
// given namespace.
func (om ObjectMap) KeysForNamespace(namespace string) (res []string) {
	for _, obj := range om[namespace] {
		key, err := cache.MetaNamespaceKeyFunc(obj)
		if err != nil {
			panic(err)
		}

		res = append(res, key)
	}

	return res
}

// AllKeys returns all namespace/name pairs in the map.
func (om ObjectMap) AllKeys() (res []string) {
	for namespace := range om {
		res = append(res, om.KeysForNamespace(namespace)...)
	}

	return res
}

// SimpleMultiIndexer is a simple map of namespace names to a cache.Indexer
// instance.  This satisfies the MultiIndexer interface.
type SimpleMultiIndexer struct {
	indexers map[string]cache.Indexer
}

// GetIndexers returns the namespace to indexer map.
func (i *SimpleMultiIndexer) GetIndexers() map[string]cache.Indexer {
	return i.indexers
}

// NewMultiIndexer returns a new SimpleMultiIndexer with indexers containing the
// data from the given ObjectMap.
func NewMultiIndexer(objects ObjectMap) *SimpleMultiIndexer {
	indexers := make(map[string]cache.Indexer, len(objects))

	for namespace := range objects {
		idx := cache.NewIndexer(
			cache.MetaNamespaceKeyFunc,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)

		for _, obj := range objects[namespace] {
			idx.Add(obj)
		}

		indexers[namespace] = idx
	}

	return &SimpleMultiIndexer{indexers: indexers}
}
