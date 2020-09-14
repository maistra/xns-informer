package informers

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/maistra/xns-informer/pkg/internal/sets"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type SimpleLister interface {
	List(selector labels.Selector) ([]runtime.Object, error)
}

type InformerFactory interface {
	Start(stopCh <-chan struct{})
	ClusterResource(resource schema.GroupVersionResource) informers.GenericInformer
	NamespacedResource(resource schema.GroupVersionResource) informers.GenericInformer
	WaitForCacheSync(stopCh <-chan struct{})
	SetNamespaces(namespaces []string)
}

type NewInformerFunc func(namespace string) informers.GenericInformer

// multiNamespaceInformerFactory provides a dynamic informer factory that
// creates informers which track changes across a set of namespaces.
type multiNamespaceInformerFactory struct {
	client       dynamic.Interface
	resyncPeriod time.Duration
	lock         sync.Mutex
	namespaces   sets.Set

	// Map of created informers by resource type.
	informers map[schema.GroupVersionResource]*multiNamespaceGenericInformer
}

var _ InformerFactory = &multiNamespaceInformerFactory{}

// NewInformerFactory returns a new factory for the given namespaces.
func NewInformerFactory(client dynamic.Interface, resync time.Duration, namespaces []string) (InformerFactory, error) {
	if len(namespaces) < 1 {
		return nil, errors.New("must provide at least one namespace")
	}

	factory := &multiNamespaceInformerFactory{
		client:       client,
		resyncPeriod: resync,
		informers:    make(map[schema.GroupVersionResource]*multiNamespaceGenericInformer),
	}

	factory.SetNamespaces(namespaces)

	return factory, nil
}

// SetNamespaces sets the list of namespaces the factory and its informers
// track.  Any new namespaces in the given set will be added to all previously
// created informers, and any namespaces that aren't in the new set will be
// removed.  You must call Start() and WaitForCacheSync() after changing the set
// of namespaces.  These are safe to call multiple times.
func (f *multiNamespaceInformerFactory) SetNamespaces(namespaces []string) {
	f.lock.Lock()
	defer f.lock.Unlock()

	newNamespaceSet := sets.NewSet(namespaces...)

	// If the set of namespaces, includes metav1.NamespaceAll, then it
	// only makes sense to create a single informer for that.
	if newNamespaceSet.Contains(metav1.NamespaceAll) {
		newNamespaceSet = sets.NewSet(metav1.NamespaceAll)
	}

	// Remove any namespaces in the current set which aren't in the
	// new set from the existing informers.
	for namespace := range f.namespaces.Difference(newNamespaceSet) {
		for _, i := range f.informers {
			i.informer.RemoveNamespace(namespace)
		}
	}

	f.namespaces = newNamespaceSet

	// Add any new namespaces to existing informers.
	for namespace := range f.namespaces {
		for _, i := range f.informers {
			i.informer.AddNamespace(namespace)
		}
	}
}

// ClusterResource returns a new cross-namespace informer for the given resource
// type and assumes it is cluster-scoped.  This means the returned informer will
// treat AddNamespace and RemoveNamespace as no-ops.
func (f *multiNamespaceInformerFactory) ClusterResource(gvr schema.GroupVersionResource) informers.GenericInformer {
	return f.ForResource(gvr, false)
}

// NamespacedResource returns a new cross-namespace informer for the given
// resource type and assumes it is namespaced.  Requesting a cluster-scoped
// resource via this method will result in errors from the underlying watch and
// will produce no events.
func (f *multiNamespaceInformerFactory) NamespacedResource(gvr schema.GroupVersionResource) informers.GenericInformer {
	return f.ForResource(gvr, true)
}

// ForResource returns a new cross-namespace informer for the given resource
// type.  If an informer for this resource type has been previously requested,
// it will be returned, otherwise a new one will be created.
//
// TODO: Should we use the discovery API to determine resource scope?
func (f *multiNamespaceInformerFactory) ForResource(gvr schema.GroupVersionResource, namespaced bool) informers.GenericInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	// Return existing informer if found.
	if informer, ok := f.informers[gvr]; ok {
		return informer
	}

	newInformerFunc := func(namespace string) informers.GenericInformer {
		// Namespace argument is ignored for cluster-scoped resources.
		if !namespaced {
			namespace = metav1.NamespaceAll
		}

		return dynamicinformer.NewFilteredDynamicInformer(
			f.client,
			gvr,
			namespace,
			f.resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			nil,
		)
	}

	informer := NewMultiNamespaceInformer(namespaced, f.resyncPeriod, newInformerFunc)
	lister := NewMultiNamespaceLister(informer, gvr)

	for namespace := range f.namespaces {
		informer.AddNamespace(namespace)
	}

	f.informers[gvr] = &multiNamespaceGenericInformer{
		informer: informer,
		lister:   lister,
	}

	return f.informers[gvr]
}

// Start starts all of the informers the factory has created to this point.
// They will be stopped when stopCh is closed.  Start is safe to call multiple
// times -- only stopped informers will be started.  This is non-blocking.
func (f *multiNamespaceInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for _, i := range f.informers {
		i.informer.NonBlockingRun(stopCh)
	}
}

// WaitForCacheSync waits for all previously started infomers caches to sync.
func (f *multiNamespaceInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) {
	syncFuncs := func() (syncFuncs []cache.InformerSynced) {
		f.lock.Lock()
		defer f.lock.Unlock()

		for _, i := range f.informers {
			syncFuncs = append(syncFuncs, i.informer.HasSynced)
		}

		return
	}

	cache.WaitForCacheSync(stopCh, syncFuncs()...)
}

// multiNamespaceGenericInformer satisfies the GenericInformer interface and
// provides cross-namespace informers and listers.
type multiNamespaceGenericInformer struct {
	informer *multiNamespaceInformer
	lister   *multiNamespaceLister
}

var _ informers.GenericInformer = &multiNamespaceGenericInformer{}

func (i *multiNamespaceGenericInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *multiNamespaceGenericInformer) Lister() cache.GenericLister {
	return i.lister
}

// informerData holds a single namespaced informer.
type informerData struct {
	informer cache.SharedIndexInformer
	lister   cache.GenericLister
	stopCh   chan struct{}
	started  bool
}

// eventHandlerData holds an event handler and its resync period.
type eventHandlerData struct {
	handler      cache.ResourceEventHandler
	resyncPeriod time.Duration
}

// multiNamespaceInformer satisfies the SharedIndexInformer interface and
// provides an informer that works across a set of namespaces -- though not all
// methods are actually usable.
type multiNamespaceInformer struct {
	informers     map[string]*informerData
	eventHandlers []eventHandlerData
	indexers      []cache.Indexers
	resyncPeriod  time.Duration
	namespaced    bool
	lock          sync.Mutex
	newInformer   NewInformerFunc
}

var _ cache.SharedIndexInformer = &multiNamespaceInformer{}

// NewMultiNamespaceInformer returns a new cross-namespace informer.  The given
// NewInformerFunc will be used to craft new single-namespace informers when
// adding namespaces.
func NewMultiNamespaceInformer(namespaced bool, resync time.Duration, newInformer NewInformerFunc) *multiNamespaceInformer {
	informer := &multiNamespaceInformer{
		informers:     make(map[string]*informerData),
		eventHandlers: make([]eventHandlerData, 0),
		indexers:      make([]cache.Indexers, 0),
		namespaced:    namespaced,
		resyncPeriod:  resync,
		newInformer:   newInformer,
	}

	// AddNamespace and RemoveNamespace are no-ops for cluster-scoped
	// informers.  They watch metav1.NamespaceAll only.
	if !namespaced {
		i := newInformer(metav1.NamespaceAll)

		informer.informers[metav1.NamespaceAll] = &informerData{
			informer: i.Informer(),
			lister:   i.Lister(),
			stopCh:   make(chan struct{}),
		}
	}

	return informer
}

func (i *multiNamespaceInformer) GetController() cache.Controller {
	panic("not implemented")
}

func (i *multiNamespaceInformer) GetStore() cache.Store {
	panic("not implemented")
}

func (i *multiNamespaceInformer) GetIndexer() cache.Indexer {
	panic("not implemented")
}

func (i *multiNamespaceInformer) LastSyncResourceVersion() string {
	panic("not implemented")
}

func (i *multiNamespaceInformer) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	panic("not implemented") // TODO: This could probably be implemented.
}

// AddNamespace adds the given namespace to the informer.  This is a no-op if an
// informer for this namespace already exists.  You must call one of the run
// functions and wait for the caches to sync before the new informer is useful.
// This is usually done via a factory with Start() and WaitForCacheSync().
func (i *multiNamespaceInformer) AddNamespace(namespace string) {
	i.lock.Lock()
	defer i.lock.Unlock()

	// If an informer for this namespace already exists, or the
	// watched resource is cluster-scoped, this is a no-op.
	if _, ok := i.informers[namespace]; ok || !i.namespaced {
		return
	}

	informer := i.newInformer(namespace)

	i.informers[namespace] = &informerData{
		informer: informer.Informer(),
		lister:   informer.Lister(),
		stopCh:   make(chan struct{}),
	}

	// Add indexers to the new informer.
	for _, idx := range i.indexers {
		i.informers[namespace].informer.AddIndexers(idx)
	}

	// Add event handlers to the new informer.
	for _, handler := range i.eventHandlers {
		i.informers[namespace].informer.AddEventHandlerWithResyncPeriod(
			handler.handler,
			handler.resyncPeriod,
		)
	}
}

// RemoveNamespace stops and deletes the informer for the given namespace.
func (i *multiNamespaceInformer) RemoveNamespace(namespace string) {
	i.lock.Lock()
	defer i.lock.Unlock()

	// If there is no informer for this namespace, or the watched
	// resource is cluster-scoped, this is a no-op.
	if _, ok := i.informers[namespace]; !ok || !i.namespaced {
		return
	}

	close(i.informers[namespace].stopCh)
	delete(i.informers, namespace)
}

// WaitForStop waits for the channel to be closed, then stops all informers.
// TODO: This may be called multiple times, but should only wait once.
func (i *multiNamespaceInformer) WaitForStop(stopCh <-chan struct{}) {
	<-stopCh // Block until stopCh is closed.
	i.lock.Lock()
	defer i.lock.Unlock()

	for _, informer := range i.informers {
		if informer.started {
			close(informer.stopCh)
			informer.started = false
		}
	}
}

// NonBlockingRun starts all stopped informers and waits for the stop channel to
// close before stopping them.  This can be called safely multiple times.
func (i *multiNamespaceInformer) NonBlockingRun(stopCh <-chan struct{}) {
	i.lock.Lock()
	defer i.lock.Unlock()

	for _, informer := range i.informers {
		if !informer.started {
			go informer.informer.Run(informer.stopCh)
			informer.started = true
		}
	}

	go i.WaitForStop(stopCh)
}

// Run starts all stopped informers and waits for the stop channel to close
// before stopping them.  This can be called safely multiple times.  This
// version blocks until the stop channel is closed.
func (i *multiNamespaceInformer) Run(stopCh <-chan struct{}) {
	i.NonBlockingRun(stopCh)
	<-stopCh // Block until stopCh is closed.
}

// AddEventHandler adds the given handler to each namespaced informer.
func (i *multiNamespaceInformer) AddEventHandler(handler cache.ResourceEventHandler) {
	i.AddEventHandlerWithResyncPeriod(handler, i.resyncPeriod)
}

// AddEventHandlerWithResyncPeriod adds the given handler with a resync period
// to each namespaced informer.  The handler will also be added to any informers
// created later as namespaces are added.
func (i *multiNamespaceInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.eventHandlers = append(i.eventHandlers, eventHandlerData{
		handler:      handler,
		resyncPeriod: resyncPeriod,
	})

	for _, informer := range i.informers {
		informer.informer.AddEventHandlerWithResyncPeriod(handler, resyncPeriod)
	}
}

// AddIndexers adds the given indexers to each namespaced informer.
func (i *multiNamespaceInformer) AddIndexers(indexers cache.Indexers) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.indexers = append(i.indexers, indexers)

	for _, informer := range i.informers {
		err := informer.informer.AddIndexers(indexers)
		if err != nil {
			return err
		}
	}

	return nil
}

// HasSynced checks if each started namespaced informer has synced.
func (i *multiNamespaceInformer) HasSynced() bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	for _, informer := range i.informers {
		if synced := informer.informer.HasSynced(); informer.started && !synced {
			return false
		}
	}

	return true
}

// getListers returns a map of namespaces to their GenericListers.
func (i *multiNamespaceInformer) getListers() map[string]cache.GenericLister {
	i.lock.Lock()
	defer i.lock.Unlock()

	res := make(map[string]cache.GenericLister, len(i.informers))
	for namespace, informer := range i.informers {
		res[namespace] = informer.lister
	}

	return res
}

// multiNamespaceLister satisfies the GenericLister interface and works across a
// set of namespaces.
type multiNamespaceLister struct {
	informer *multiNamespaceInformer
	resource schema.GroupVersionResource
}

var _ cache.GenericLister = &multiNamespaceLister{}

// NewMultiNamespaceLister returns a new cross-namespace lister.
func NewMultiNamespaceLister(i *multiNamespaceInformer, gvr schema.GroupVersionResource) *multiNamespaceLister {
	return &multiNamespaceLister{informer: i, resource: gvr}
}

// List returns all objects matching the given label selector across all
// namespaces the lister's backing informer knows about.  The resulting objects
// will be runtime.Object interfaces backed by unstructured types.  Use the
// conversion methods if you want concrete types.
func (l *multiNamespaceLister) List(selector labels.Selector) ([]runtime.Object, error) {
	var result []runtime.Object

	for _, lister := range l.informer.getListers() {
		objs, err := lister.List(selector)
		if err != nil {
			return nil, err
		}

		result = append(result, objs...)
	}

	return result, nil
}

// Get fetches the named object from the backing informer's cache.  As with
// List(), the returned object is a runtime.Object interface backed by an
// unstructured object.  Use the conversion methods if you want concrete types.
func (l *multiNamespaceLister) Get(name string) (runtime.Object, error) {
	namespace, _, err := cache.SplitMetaNamespaceKey(name)
	if err != nil {
		return nil, err
	}

	return l.ByNamespace(namespace).Get(name)
}

// ByNamespace returns a GenericNamespaceLister for the given namespace.  If the
// backing informer doesn't track this namespace, a dummy lister that always
// returns a not-found error is returned.
func (l *multiNamespaceLister) ByNamespace(namespace string) cache.GenericNamespaceLister {
	listers := l.informer.getListers()

	if lister, ok := listers[metav1.NamespaceAll]; ok {
		return lister.ByNamespace(namespace)
	} else if lister, ok := listers[namespace]; ok {
		return lister.ByNamespace(namespace)
	}

	return &NilNamespaceLister{
		namespace: namespace,
		resource:  l.resource,
	}
}

// NilNamespaceLister is a GenericNamespaceLister that always returns not-found
// errors.  It is returned when a multiNamespaceLister is asked for a namespaced
// lister for a namespace it doesn't know about.
type NilNamespaceLister struct {
	namespace string
	resource  schema.GroupVersionResource
}

var _ cache.GenericNamespaceLister = &NilNamespaceLister{}

func (l *NilNamespaceLister) error(name string) error {
	return &apierrors.StatusError{
		ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusNotFound,
			Reason: metav1.StatusReasonNotFound,
			Details: &metav1.StatusDetails{
				Group: l.resource.Group,
				Kind:  l.resource.Resource,
				Name:  name,
			},
			Message: fmt.Sprintf("namespace %q not included in informer cache", l.namespace),
		},
	}
}

func (l *NilNamespaceLister) List(selector labels.Selector) ([]runtime.Object, error) {
	return nil, l.error("")
}

func (l *NilNamespaceLister) Get(name string) (runtime.Object, error) {
	return nil, l.error(name)
}

// ConvertUnstructured takes a runtime.Object, which *must* be backed by a
// pointer to an unstructured object, and a second runtime.Object which should
// be backed by a concrete type that the unstructured object is expected to
// represent.  ConvertUnstructured will use the default converter from the
// runtime package to convert the unstructured object to the concrete type.
func ConvertUnstructured(unstructuredObj runtime.Object, out runtime.Object) error {
	u, ok := unstructuredObj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured conversion failed")
	}

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, out)
	if err != nil {
		return err
	}

	return nil
}
