package testing

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// FakeCreationTimestamp is a dummy creation timestamp for test objects.
var FakeCreationTimestamp = metav1.NewTime(time.Unix(1601548012, 0))

// ConfigMapGVR is the GroupVersionResource for v1 ConfigMaps.
var ConfigMapGVR = corev1.SchemeGroupVersion.WithResource("configmaps")

// NewConfigMap returns a ConfigMap with the given namespace, name, and data.
func NewConfigMap(ns, name string, data map[string]string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: FakeCreationTimestamp,
			Namespace:         ns,
			Name:              name,
		},
		Data: data,
	}

	return cm
}

// NewUnstructuredConfigMap returns an unstructured.Unstructured object
// representing a ConfigMap with the given namespace, name, and data.
func NewUnstructuredConfigMap(ns, name string, data map[string]string) *unstructured.Unstructured {
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
				"creationTimestamp": FakeCreationTimestamp.ToUnstructured(),
			},
			"data": unstructuredData,
		},
	}
}
