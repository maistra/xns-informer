package testing

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	kubetesting "k8s.io/client-go/testing"
)

var (
	timestamp    = metav1.NewTime(time.Now().UTC())
	configMapGVR = corev1.SchemeGroupVersion.WithResource("configmaps")
)

func newConfigMap(ns, name string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         ns,
			CreationTimestamp: timestamp,
		},
		Data: data,
	}
}

func newUnstructuredConfigMap(ns, name string, data map[string]string) *unstructured.Unstructured {
	unstructuredData := make(map[string]interface{}, len(data))

	for k, v := range data {
		unstructuredData[k] = v
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "ConfigMap",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"name":              name,
				"namespace":         ns,
				"creationTimestamp": timestamp.ToUnstructured(),
			},
			"data": unstructuredData,
		},
	}
}

func TestObjectsToUnstructured(t *testing.T) {
	testData := map[string]string{"foo": "bar"}

	testCases := []struct {
		name        string
		converter   runtime.ObjectConvertor
		in          []runtime.Object
		out         []runtime.Object
		errExpected bool
	}{
		{
			name:      "no objects",
			converter: scheme.Scheme,
			in:        []runtime.Object{},
			out:       []runtime.Object{},
		},
		{
			name:      "some objects",
			converter: scheme.Scheme,
			in: []runtime.Object{
				newConfigMap("test-ns", "test-cm-1", testData),
				newConfigMap("test-ns", "test-cm-2", testData),
			},
			out: []runtime.Object{
				newUnstructuredConfigMap("test-ns", "test-cm-1", testData),
				newUnstructuredConfigMap("test-ns", "test-cm-2", testData),
			},
		},
		{
			name:      "empty scheme",
			converter: runtime.NewScheme(),
			in: []runtime.Object{
				newConfigMap("test-ns", "test-cm", testData),
			},
			errExpected: true, // Conversion should fail.
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			objs, err := ObjectsToUnstructured(tt.converter, tt.in...)
			if err != nil {
				if tt.errExpected {
					return
				}

				t.Fatalf("ObjectsToUnstructured failed: %v", err)
			}

			for i, obj := range objs {
				if !equality.Semantic.DeepEqual(obj, tt.out[i]) {
					t.Fatalf("\n- got: %#v\n- want: %#v", obj, tt.out[i])
				}
			}
		})
	}
}

func TestUnstructuredObjectReflector(t *testing.T) {
	testData := map[string]string{"foo": "bar"}
	updatedData := map[string]string{"foo": "updated"}
	cmJSONPatch := `[{"op":"replace","path":"/data/foo","value":"updated"}]`

	testCases := []struct {
		name     string
		scheme   *runtime.Scheme
		actions  []kubetesting.Action
		expected []*unstructured.Unstructured
	}{
		{
			// A simple create.
			name:   "create",
			scheme: scheme.Scheme,
			actions: []kubetesting.Action{
				kubetesting.CreateActionImpl{
					ActionImpl: kubetesting.ActionImpl{
						Namespace: "test-ns",
						Verb:      "create",
						Resource:  configMapGVR,
					},
					Object: newConfigMap("test-ns", "test-cm", testData),
				},
			},
			expected: []*unstructured.Unstructured{
				newUnstructuredConfigMap("test-ns", "test-cm", testData),
			},
		},
		{
			// A create followed by an update.
			name:   "update",
			scheme: scheme.Scheme,
			actions: []kubetesting.Action{
				kubetesting.CreateActionImpl{
					ActionImpl: kubetesting.ActionImpl{
						Namespace: "test-ns",
						Verb:      "create",
						Resource:  configMapGVR,
					},
					Object: newConfigMap("test-ns", "test-cm", testData),
				},
				kubetesting.UpdateActionImpl{
					ActionImpl: kubetesting.ActionImpl{
						Namespace: "test-ns",
						Verb:      "update",
						Resource:  configMapGVR,
					},
					Object: newConfigMap("test-ns", "test-cm", updatedData),
				},
			},
			expected: []*unstructured.Unstructured{
				newUnstructuredConfigMap("test-ns", "test-cm", updatedData),
			},
		},
		{
			// A create followed by a delete.
			name:   "delete",
			scheme: scheme.Scheme,
			actions: []kubetesting.Action{
				kubetesting.CreateActionImpl{
					ActionImpl: kubetesting.ActionImpl{
						Namespace: "test-ns",
						Verb:      "create",
						Resource:  configMapGVR,
					},
					Object: newConfigMap("test-ns", "test-cm", testData),
				},
				kubetesting.DeleteActionImpl{
					ActionImpl: kubetesting.ActionImpl{
						Namespace: "test-ns",
						Verb:      "delete",
						Resource:  configMapGVR,
					},
					Name: "test-cm",
				},
			},
			expected: []*unstructured.Unstructured{},
		},
		{
			// A create followed by a patch.
			name:   "patch",
			scheme: scheme.Scheme,
			actions: []kubetesting.Action{
				kubetesting.CreateActionImpl{
					ActionImpl: kubetesting.ActionImpl{
						Namespace: "test-ns",
						Verb:      "create",
						Resource:  configMapGVR,
					},
					Object: newConfigMap("test-ns", "test-cm", testData),
				},
				kubetesting.PatchActionImpl{
					ActionImpl: kubetesting.ActionImpl{
						Namespace: "test-ns",
						Verb:      "patch",
						Resource:  configMapGVR,
					},
					Name:      "test-cm",
					PatchType: types.JSONPatchType,
					Patch:     []byte(cmJSONPatch),
				},
			},
			expected: []*unstructured.Unstructured{
				newUnstructuredConfigMap("test-ns", "test-cm", updatedData),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dynamicClient := dynamicfake.NewSimpleDynamicClient(tt.scheme)
			reflector := UnstructuredObjectReflector(tt.scheme, &dynamicClient.Fake)

			// Execute each action through the reflector.
			for _, action := range tt.actions {
				reflector(action)
			}

			objList, err := dynamicClient.Resource(configMapGVR).Namespace("test-ns").
				List(context.TODO(), metav1.ListOptions{})

			if err != nil {
				t.Fatalf("List failed: %v", err)
			}

			// Ensure the expected number of objects exist.
			if len(objList.Items) != len(tt.expected) {
				t.Fatalf("Found %d items, but expected %d.",
					len(objList.Items), len(tt.expected))
			}

			// Ensure each expected object is found.
			for _, expectedObj := range tt.expected {
				name := expectedObj.GetName()
				ns := expectedObj.GetNamespace()

				gotObj, err := dynamicClient.Resource(configMapGVR).Namespace(ns).
					Get(context.TODO(), name, metav1.GetOptions{})

				if err != nil {
					t.Fatalf("Get failed: %v", err)
				}

				if !equality.Semantic.DeepEqual(gotObj, expectedObj) {
					t.Fatalf("\n- got: %#v\n- want: %#v", gotObj, expectedObj)
				}
			}
		})
	}
}

func TestCreateNewFakeClients(t *testing.T) {
	kc, dc, err := NewFakeClients(scheme.Scheme)
	if err != nil {
		t.Fatalf("Failed to create clients: %v", err)
	}

	ns := "test-ns"
	name := "test-cm"
	ctx := context.TODO()

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Data: map[string]string{"test": "data"},
	}

	// Create ConfigMap with typed client.
	_, err = kc.CoreV1().ConfigMaps(ns).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create ConfigMap: %v", err)
	}

	// Fetch typed ConfigMap with typed client.
	typed, err := kc.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to fetch typed ConfigMap: %v", err)
	}

	// Fetch unstructured ConfigMap with dynamic client.
	u, err := dc.Resource(configMapGVR).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to fetch unstructured ConfigMap: %v", err)
	}

	converted := &corev1.ConfigMap{}
	if err := scheme.Scheme.Convert(u, converted, nil); err != nil {
		t.Fatalf("Failed to convert unstructured ConfigMap: %v", err)
	}

	if !equality.Semantic.DeepEqual(typed, converted) {
		t.Fatalf("Fetched ConfigMaps not equal!\n%#v\n%#v\n", typed, converted)
	}
}
