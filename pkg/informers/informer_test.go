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

	internaltesting "github.com/maistra/xns-informer/pkg/internal/testing"
)

const (
	resyncPeriod = 0
)

var configMapGVR = corev1.SchemeGroupVersion.WithResource("configmaps")

func TestEventHandlers(t *testing.T) {
	var err error

	ctx := context.TODO()
	stopCh := make(chan struct{})

	namespaces := []string{"ns1", "ns2"}

	cm1 := internaltesting.NewConfigMap(namespaces[0], "cm1", nil)
	cm2 := internaltesting.NewConfigMap(namespaces[1], "cm2", nil)

	addFuncCalled := false
	updateFuncCalled := false
	deleteFuncCalled := false

	client := kubefake.NewSimpleClientset()
	namespaceSet := NewNamespaceSet(namespaces...)

	informer := NewMultiNamespaceInformer(namespaceSet, 0, func(namespace string) cache.SharedIndexInformer {
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

	// Create the ConfigMap in the first namespace.
	_, err = client.CoreV1().ConfigMaps(namespaces[0]).Create(ctx, cm1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create ConfigMap: %v", err)
	}

	// Create the ConfigMap in the second namespace.
	_, err = client.CoreV1().ConfigMaps(namespaces[1]).Create(ctx, cm2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create ConfigMap: %v", err)
	}

	// Update the ConfigMap in the first namespace.
	_, err = client.CoreV1().ConfigMaps(namespaces[0]).Update(ctx, cm1, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("Failed to update ConfigMap: %v", err)
	}

	// Delete the ConfigMap in the second namespace.
	err = client.CoreV1().ConfigMaps(namespaces[1]).Delete(ctx, "cm2", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Failed to delete ConfigMap: %v", err)
	}

	// Wait for all handler functions to be called.
	err = wait.PollImmediate(100*time.Millisecond, 1*time.Minute, func() (bool, error) {
		return addFuncCalled && updateFuncCalled && deleteFuncCalled, nil
	})

	if err != nil {
		t.Fatalf("Handlers not called: %v", err)
	}

	// Remove first namespace from the set.
	deleteFuncCalled = false
	informer.RemoveNamespace(namespaces[0])

	// Wait for delete handler function to be called again.
	err = wait.PollImmediate(100*time.Millisecond, 1*time.Minute, func() (bool, error) {
		return deleteFuncCalled, nil
	})

	if err != nil {
		t.Fatalf("Delete handler not called after namespace removal: %v", err)
	}
}
