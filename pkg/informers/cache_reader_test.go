package informers_test

import (
	"sort"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	internaltesting "github.com/maistra/xns-informer/pkg/internal/testing"
)

var cacheReaderTests = []struct {
	name    string
	objects internaltesting.ObjectMap
}{
	{
		name:    "empty",
		objects: nil,
	},
	{
		name: "single-namespace",
		objects: internaltesting.ObjectMap{
			"namespace-one": {
				internaltesting.NewConfigMap("namespace-one", "cm-one-one", nil),
				internaltesting.NewConfigMap("namespace-one", "cm-one-two", nil),
			},
		},
	},
	{
		name: "multi-namespace",
		objects: internaltesting.ObjectMap{
			"namespace-one": {internaltesting.NewConfigMap("namespace-one", "cm-one", nil)},
			"namespace-two": {internaltesting.NewConfigMap("namespace-two", "cm-two", nil)},
		},
	},
}

func TestCacheReaderList(t *testing.T) {
	for _, tt := range cacheReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			idx := internaltesting.NewMultiIndexer(tt.objects)
			reader := xnsinformers.NewCacheReader(idx)

			gotObjects := reader.List()
			expectedObjects := tt.objects.Objects()

			// Convert to runtime.Object for comparison.
			runtimeObjects := make([]runtime.Object, len(gotObjects))
			for i, obj := range gotObjects {
				runtimeObjects[i] = obj.(*corev1.ConfigMap)
			}

			internaltesting.SortRuntimeObjects(runtimeObjects)
			internaltesting.SortRuntimeObjects(expectedObjects)

			if !equality.Semantic.DeepEqual(runtimeObjects, expectedObjects) {
				t.Errorf("\n- got: %#v\n- want: %#v", runtimeObjects, expectedObjects)
			}
		})
	}
}

func TestCacheReaderListKeys(t *testing.T) {
	for _, tt := range cacheReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			idx := internaltesting.NewMultiIndexer(tt.objects)
			reader := xnsinformers.NewCacheReader(idx)

			gotKeys := reader.ListKeys()
			expectedKeys := tt.objects.AllKeys()

			sort.Strings(gotKeys)
			sort.Strings(expectedKeys)
			if !equality.Semantic.DeepEqual(gotKeys, expectedKeys) {
				t.Errorf("\n- got: %#v\n- want: %#v", gotKeys, expectedKeys)
			}
		})
	}
}

func TestCacheReaderGetIndexers(t *testing.T) {
	for _, tt := range cacheReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			idx := internaltesting.NewMultiIndexer(tt.objects)
			reader := xnsinformers.NewCacheReader(idx)

			if len(tt.objects) == 0 {
				t.SkipNow()
			}

			// Check that we have the expected single namespace indexer.
			gotIndexers := reader.GetIndexers()
			if len(gotIndexers) != 1 || gotIndexers[cache.NamespaceIndex] == nil {
				t.Errorf("Wrong number of indexers or missing namespace indexer!")
			}

		})
	}
}

func TestCacheReaderIndex(t *testing.T) {
	for _, tt := range cacheReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			idx := internaltesting.NewMultiIndexer(tt.objects)
			reader := xnsinformers.NewCacheReader(idx)

			// Use the namespace indexer to fetch all objects in a particular
			// namespace and compare that to the input.
			namespace := "namespace-one"
			objMeta := &metav1.ObjectMeta{Namespace: namespace}
			expectedObjects := tt.objects[namespace]
			gotObjects, err := reader.Index(cache.NamespaceIndex, objMeta)
			if err != nil {
				t.Fatalf("Error fetching by index: %v", err)
			}

			// Convert to runtime.Object for comparison.
			runtimeObjects := make([]runtime.Object, len(gotObjects))
			for i, obj := range gotObjects {
				runtimeObjects[i] = obj.(*corev1.ConfigMap)
			}

			internaltesting.SortRuntimeObjects(runtimeObjects)
			internaltesting.SortRuntimeObjects(expectedObjects)

			if !equality.Semantic.DeepEqual(runtimeObjects, expectedObjects) {
				t.Errorf("\n- got: %#v\n- want: %#v", runtimeObjects, expectedObjects)
			}
		})
	}
}

func TestCacheReaderIndexKeys(t *testing.T) {
	for _, tt := range cacheReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			idx := internaltesting.NewMultiIndexer(tt.objects)
			reader := xnsinformers.NewCacheReader(idx)

			// Test IndexKeys()
			namespace := "namespace-two"
			expectedKeys := tt.objects.KeysForNamespace(namespace)
			gotKeys, err := reader.IndexKeys(cache.NamespaceIndex, namespace)
			if err != nil {
				t.Errorf("Error fetching index keys: %v", err)
			}

			sort.Strings(gotKeys)
			sort.Strings(expectedKeys)
			if !equality.Semantic.DeepEqual(gotKeys, expectedKeys) {
				t.Errorf("\n- got: %#v\n- want: %#v", gotKeys, expectedKeys)
			}
		})
	}
}

func TestCacheReaderListIndexFuncValues(t *testing.T) {
	for _, tt := range cacheReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			idx := internaltesting.NewMultiIndexer(tt.objects)
			reader := xnsinformers.NewCacheReader(idx)

			gotValues := reader.ListIndexFuncValues(cache.NamespaceIndex)
			expectedValues := tt.objects.Namespaces()

			sort.Strings(gotValues)
			sort.Strings(expectedValues)
			if !equality.Semantic.DeepEqual(gotValues, expectedValues) {
				t.Errorf("\n- got: %#v\n- want: %#v", gotValues, expectedValues)
			}
		})
	}
}

func TestCacheReaderByIndex(t *testing.T) {
	for _, tt := range cacheReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			idx := internaltesting.NewMultiIndexer(tt.objects)
			reader := xnsinformers.NewCacheReader(idx)

			for _, namespace := range tt.objects.Namespaces() {
				expectedObjects := tt.objects[namespace]
				gotObjects, err := reader.ByIndex(cache.NamespaceIndex, namespace)
				if err != nil {
					t.Fatalf("Error fetching by index: %v", err)
				}

				// Convert to runtime.Object for comparison.
				runtimeObjects := make([]runtime.Object, len(gotObjects))
				for i, obj := range gotObjects {
					runtimeObjects[i] = obj.(*corev1.ConfigMap)
				}

				internaltesting.SortRuntimeObjects(runtimeObjects)
				internaltesting.SortRuntimeObjects(expectedObjects)

				if !equality.Semantic.DeepEqual(runtimeObjects, expectedObjects) {
					t.Errorf("\n- got: %#v\n- want: %#v", runtimeObjects, expectedObjects)
				}
			}

			// Bad index name should return an error.
			_, err := reader.ByIndex("bad-index", "missing")
			if len(tt.objects) != 0 && err == nil {
				t.Error("Expected error but got nil")
			}

			// Bad object name should return an empty result.
			res, err := reader.ByIndex(cache.NamespaceIndex, "missing")
			if err != nil {
				t.Fatalf("Error fetching by index: %v", err)
			} else if len(res) != 0 {
				t.Errorf("Result length should be 0 not %d", len(res))
			}
		})
	}
}

func TestCacheReaderGet(t *testing.T) {
	for _, tt := range cacheReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			idx := internaltesting.NewMultiIndexer(tt.objects)
			reader := xnsinformers.NewCacheReader(idx)

			for _, expectedObject := range tt.objects.Objects() {
				gotObject, exists, err := reader.Get(expectedObject)
				if err != nil {
					t.Fatalf("Error fetching object: %v", err)
				} else if !exists {
					t.Fatalf("Expected object not found")
				}

				runtimeObject, ok := gotObject.(runtime.Object)
				if !ok {
					t.Fatalf("%T is not a runtime.Object", gotObject)
				}

				if !equality.Semantic.DeepEqual(runtimeObject, expectedObject) {
					t.Errorf("\n- got: %#v\n- want: %#v", runtimeObject, expectedObject)
				}
			}

			// Look for an object that should not exist.
			missingObject := &metav1.ObjectMeta{Namespace: "missing", Name: "missing"}
			_, exists, err := reader.Get(missingObject)
			if err != nil {
				t.Fatalf("Error fetching object: %v", err)
			} else if exists {
				t.Error("Found unexpected object")
			}

			// This should return an error.
			_, _, err = reader.Get(nil)
			if err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}

func TestCacheReaderGetByKey(t *testing.T) {
	for _, tt := range cacheReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			idx := internaltesting.NewMultiIndexer(tt.objects)
			reader := xnsinformers.NewCacheReader(idx)

			for _, expectedObject := range tt.objects.Objects() {
				key, _ := cache.MetaNamespaceKeyFunc(expectedObject)
				gotObject, exists, err := reader.GetByKey(key)
				if err != nil {
					t.Fatalf("Error fetching object: %v", err)
				} else if !exists {
					t.Fatalf("Expected object not found")
				}

				runtimeObject, ok := gotObject.(runtime.Object)
				if !ok {
					t.Fatalf("%T is not a runtime.Object", gotObject)
				}

				if !equality.Semantic.DeepEqual(runtimeObject, expectedObject) {
					t.Errorf("\n- got: %#v\n- want: %#v", runtimeObject, expectedObject)
				}
			}
		})
	}
}
