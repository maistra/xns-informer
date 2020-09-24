package testing

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestCreate(t *testing.T) {
	s := scheme.Scheme
	kc, dc, err := NewFakeClients(s)
	if err != nil {
		t.Fatalf("Failed to create clients: %v", err)
	}

	gvr := corev1.SchemeGroupVersion.WithResource("configmaps")

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "test-ns",
		},
		Data: map[string]string{
			"test": "test data",
		},
	}

	_, err = kc.CoreV1().ConfigMaps("test-ns").Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create ConfigMap: %v", err)
	}

	typedCM, err := kc.CoreV1().ConfigMaps("test-ns").
		Get(context.TODO(), "test-configmap", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to fetch typed ConfigMap: %v", err)
	}

	unstructuredCM, err := dc.Resource(gvr).Namespace("test-ns").
		Get(context.TODO(), "test-configmap", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to fetch unstructured ConfigMap: %v", err)
	}

	convertedCM := &corev1.ConfigMap{}
	if err := s.Convert(unstructuredCM, convertedCM, nil); err != nil {
		t.Fatalf("Failed to convert unstructured ConfigMap: %v", err)
	}

	if !equality.Semantic.DeepEqual(typedCM, convertedCM) {
		t.Fatal("Fetched ConfigMaps not equal!")
	}
}
