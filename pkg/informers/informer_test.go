package informers

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

const (
	resyncPeriod = 0
)

var configMapGVR = corev1.SchemeGroupVersion.WithResource("configmaps")

func TestEventHandlers(t *testing.T) {
	var err error

	ctx := context.TODO()
	stopCh := make(chan struct{})

	ns := "test-ns"
	name := "test-configmap"

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Data: map[string]string{"test": "data"},
	}

	addFuncCalled := false
	updateFuncCalled := false
	deleteFuncCalled := false

	client := kubefake.NewSimpleClientset()
	namespaces := NewNamespaceSet(ns)

	informer := NewMultiNamespaceInformer(namespaces, 0, func(namespace string) cache.SharedIndexInformer {
		return cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					return client.CoreV1().ConfigMaps(namespace).List(ctx, options)
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return client.CoreV1().ConfigMaps(namespace).Watch(ctx, options)
				},
			},
			&corev1.ConfigMap{},
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)
	})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(_ interface{}) {
			addFuncCalled = true
		},
		UpdateFunc: func(_, _ interface{}) {
			updateFuncCalled = true
		},
		DeleteFunc: func(_ interface{}) {
			deleteFuncCalled = true
		},
	})

	go informer.Run(stopCh)
	cache.WaitForCacheSync(stopCh, informer.HasSynced)

	// Create the ConfigMap.
	_, err = client.CoreV1().ConfigMaps(ns).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create ConfigMap: %v", err)
	}

	// Update the ConfigMap.
	_, err = client.CoreV1().ConfigMaps(ns).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update ConfigMap: %v", err)
	}

	// Delete the ConfigMap.
	err = client.CoreV1().ConfigMaps(ns).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Failed to delete ConfigMap: %v", err)
	}

	// Wait for all handler functions to be called.
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		return addFuncCalled && updateFuncCalled && deleteFuncCalled, nil
	})

	if err != nil {
		t.Fatalf("Handler not called: %v", err)
	}
}
