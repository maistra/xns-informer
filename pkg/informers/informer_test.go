package informers_test

// These are largely adapted from the upstream tests for shared informers, with
// a few changes to help test multiple namespaces.

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	fcache "k8s.io/client-go/tools/cache/testing"

	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	internaltesting "github.com/maistra/xns-informer/pkg/internal/testing"
)

type testListener struct {
	lock              sync.RWMutex
	resyncPeriod      time.Duration
	expectedItemNames sets.String
	receivedItemNames []string
	name              string
}

func newTestListener(name string, resyncPeriod time.Duration, expected ...string) *testListener {
	l := &testListener{
		resyncPeriod:      resyncPeriod,
		expectedItemNames: sets.NewString(expected...),
		name:              name,
	}
	return l
}

func (l *testListener) OnAdd(obj interface{}) {
	l.handle(obj)
}

func (l *testListener) OnUpdate(old, new interface{}) {
	l.handle(new)
}

func (l *testListener) OnDelete(obj interface{}) {
}

func (l *testListener) handle(obj interface{}) {
	key, _ := cache.MetaNamespaceKeyFunc(obj)
	fmt.Printf("%s: handle: %v\n", l.name, key)
	l.lock.Lock()
	defer l.lock.Unlock()

	objectMeta, _ := meta.Accessor(obj)
	l.receivedItemNames = append(l.receivedItemNames, objectMeta.GetName())
}

func (l *testListener) ok() bool {
	fmt.Println("polling")
	err := wait.PollImmediate(100*time.Millisecond, 2*time.Second, func() (bool, error) {
		if l.satisfiedExpectations() {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return false
	}

	// wait just a bit to allow any unexpected stragglers to come in
	fmt.Println("sleeping")
	time.Sleep(1 * time.Second)
	fmt.Println("final check")
	return l.satisfiedExpectations()
}

func (l *testListener) satisfiedExpectations() bool {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return sets.NewString(l.receivedItemNames...).Equal(l.expectedItemNames)
}

// create a new NewMultiNamespaceInformer with the given example object and map
// of namespaces to ListWatchers, which are usually a FakeControllerSource here.
func newInformer(obj runtime.Object, lws map[string]cache.ListerWatcher) xnsinformers.MultiNamespaceInformer {
	resync := 1 * time.Second
	namespaceSet := xnsinformers.NewNamespaceSet()
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}

	var namespaces []string
	for ns := range lws {
		namespaces = append(namespaces, ns)
	}

	namespaceSet.SetNamespaces(namespaces)

	return xnsinformers.NewMultiNamespaceInformer(namespaceSet, resync, func(ns string) cache.SharedIndexInformer {
		return cache.NewSharedIndexInformer(lws[ns], obj, resync, indexers)
	})
}

// verify that https://github.com/kubernetes/kubernetes/issues/59822 is fixed
func TestSharedInformerInitializationRace(t *testing.T) {
	source := fcache.NewFakeControllerSource()
	informer := newInformer(&v1.Pod{}, map[string]cache.ListerWatcher{
		"ns1": source,
	})
	listener := newTestListener("raceListener", 0)

	stop := make(chan struct{})
	go informer.AddEventHandlerWithResyncPeriod(listener, listener.resyncPeriod)
	go informer.Run(stop)
	close(stop)
}

// TestSharedInformerWatchDisruption simulates a watch that was closed
// with updates to the store during that time. We ensure that handlers with
// resync and no resync see the expected state.
func TestSharedInformerWatchDisruption(t *testing.T) {
	// source simulates an apiserver object endpoint.
	source := fcache.NewFakeControllerSource()

	source.Add(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", UID: "pod1", ResourceVersion: "1"}})
	source.Add(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", UID: "pod2", ResourceVersion: "2"}})

	// create the shared informer and resync every 1s
	informer := newInformer(&v1.Pod{}, map[string]cache.ListerWatcher{
		"ns1": source,
	})

	// listener, never resync
	listenerNoResync := newTestListener("listenerNoResync", 0, "pod1", "pod2")
	informer.AddEventHandlerWithResyncPeriod(listenerNoResync, listenerNoResync.resyncPeriod)

	listenerResync := newTestListener("listenerResync", 1*time.Second, "pod1", "pod2")
	informer.AddEventHandlerWithResyncPeriod(listenerResync, listenerResync.resyncPeriod)
	listeners := []*testListener{listenerNoResync, listenerResync}

	stop := make(chan struct{})
	defer close(stop)

	go informer.Run(stop)

	for _, listener := range listeners {
		if !listener.ok() {
			t.Errorf("%s: expected %v, got %v", listener.name, listener.expectedItemNames, listener.receivedItemNames)
		}
	}

	// Add pod3, bump pod2 but don't broadcast it, so that the change will be seen only on relist
	source.AddDropWatch(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod3", UID: "pod3", ResourceVersion: "3"}})
	source.ModifyDropWatch(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", UID: "pod2", ResourceVersion: "4"}})

	// Ensure that nobody saw any changes
	for _, listener := range listeners {
		if !listener.ok() {
			t.Errorf("%s: expected %v, got %v", listener.name, listener.expectedItemNames, listener.receivedItemNames)
		}
	}

	for _, listener := range listeners {
		listener.lock.Lock()
		listener.receivedItemNames = []string{}
		listener.lock.Unlock()
	}

	listenerNoResync.expectedItemNames = sets.NewString("pod2", "pod3")
	listenerResync.expectedItemNames = sets.NewString("pod1", "pod2", "pod3")

	// This calls shouldSync, which deletes noResync from the list of syncingListeners
	time.Sleep(1 * time.Second)

	// Simulate a connection loss (or even just a too-old-watch)
	source.ResetWatch()

	for _, listener := range listeners {
		if !listener.ok() {
			t.Errorf("%s: expected %v, got %v", listener.name, listener.expectedItemNames, listener.receivedItemNames)
		}
	}
}

func TestSharedInformerErrorHandling(t *testing.T) {
	source1 := fcache.NewFakeControllerSource()
	source1.Add(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}})
	source1.ListError = fmt.Errorf("Access Denied")

	source2 := fcache.NewFakeControllerSource()
	source2.Add(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}})

	informer := newInformer(&v1.Pod{}, map[string]cache.ListerWatcher{
		"ns1": source1,
		"ns2": source2,
	})

	// ns1 source should throw an error.
	errCh := make(chan error)
	_ = informer.SetWatchErrorHandler(func(_ *cache.Reflector, err error) {
		errCh <- err
	})

	// ns2 source should succeed.
	ns2Listener := newTestListener("ns2Listener", 0, "pod2")
	informer.AddEventHandlerWithResyncPeriod(ns2Listener, ns2Listener.resyncPeriod)

	stop := make(chan struct{})
	go informer.Run(stop)

	select {
	case err := <-errCh:
		if !strings.Contains(err.Error(), "Access Denied") {
			t.Errorf("Expected 'Access Denied' error. Actual: %v", err)
		}
	case <-time.After(time.Second):
		t.Errorf("Timeout waiting for error handler call")
	}
	close(stop)

	if !ns2Listener.ok() {
		t.Errorf("%s: expected %v, got %v",
			ns2Listener.name,
			ns2Listener.expectedItemNames,
			ns2Listener.receivedItemNames,
		)
	}
}

func TestMultiNamespaceInformerEventHandlers(t *testing.T) {
	var err error

	ctx := context.TODO()
	stopCh := make(chan struct{})

	namespaces := []string{"ns1", "ns2"}

	cm1 := internaltesting.NewConfigMap(namespaces[0], "cm1", nil)
	cm2 := internaltesting.NewConfigMap(namespaces[1], "cm2", nil)

	lock := sync.RWMutex{}
	addFuncCalled := false
	updateFuncCalled := false
	deleteFuncCalled := false

	// These tests use the fake client instead of a FakeControllerSource.
	client := kubefake.NewSimpleClientset()
	namespaceSet := xnsinformers.NewNamespaceSet(namespaces...)

	informer := xnsinformers.NewMultiNamespaceInformer(namespaceSet, 0, func(namespace string) cache.SharedIndexInformer {
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
			0,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)
	})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(_ interface{}) {
			lock.Lock()
			addFuncCalled = true
			lock.Unlock()
		},
		UpdateFunc: func(_, _ interface{}) {
			lock.Lock()
			updateFuncCalled = true
			lock.Unlock()
		},
		DeleteFunc: func(_ interface{}) {
			lock.Lock()
			deleteFuncCalled = true
			lock.Unlock()
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
		lock.RLock()
		defer lock.RUnlock()
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

func TestMultiNamespaceInformerHasSynced(t *testing.T) {
	namespaceSet := xnsinformers.NewNamespaceSet()
	hasSynced := false

	informer := xnsinformers.NewMultiNamespaceInformer(namespaceSet, 0, func(namespace string) cache.SharedIndexInformer {
		return mockInformer{
			hasSynced: &hasSynced,
		}
	})

	if informer.HasSynced() {
		t.Fatalf("informer is synced, but shouldn't be because namespaces haven't been set yet")
	}

	namespaceSet.SetNamespaces([]string{"ns1", "ns2"})

	if informer.HasSynced() {
		t.Fatalf("informer is synced, but shouldn't be because the underlying informers aren't synced")
	}

	hasSynced = true

	if !informer.HasSynced() {
		t.Fatalf("expected informer to be synced")
	}
}

type mockInformer struct {
	hasSynced *bool
}

func (m mockInformer) AddEventHandler(handler cache.ResourceEventHandler) {
	panic("not implemented")
}

func (m mockInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	panic("not implemented")
}

func (m mockInformer) GetStore() cache.Store {
	panic("not implemented")
}

func (m mockInformer) GetController() cache.Controller {
	panic("not implemented")
}

func (m mockInformer) Run(stopCh <-chan struct{}) {
	panic("not implemented")
}

func (m mockInformer) HasSynced() bool {
	return *m.hasSynced
}

func (m mockInformer) LastSyncResourceVersion() string {
	panic("not implemented")
}

func (m mockInformer) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	panic("not implemented")
}

func (m mockInformer) AddIndexers(indexers cache.Indexers) error {
	panic("not implemented")
}

func (m mockInformer) GetIndexer() cache.Indexer {
	panic("not implemented")
}
