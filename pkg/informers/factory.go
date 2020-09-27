package informers

import (
	"sync"
	"time"

	"github.com/maistra/xns-informer/pkg/internal/sets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/dynamiclister"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
)

// SharedInformerFactory provides shared informers for any resource type and
// works across a set of namespaces, which can be updated at any time.
type SharedInformerFactory interface {
	Start(stopCh <-chan struct{})
	ClusterResource(resource schema.GroupVersionResource) informers.GenericInformer
	NamespacedResource(resource schema.GroupVersionResource) informers.GenericInformer
	WaitForCacheSync(stopCh <-chan struct{})
	SetNamespaces(namespaces []string)
	GetScheme() *runtime.Scheme
}

// SharedInformerOption is a functional option for a SharedInformerFactory.
type SharedInformerOption func(*multiNamespaceInformerFactory) *multiNamespaceInformerFactory

// WithScheme sets a custom scheme for a SharedInformerFactory.
func WithScheme(scheme *runtime.Scheme) SharedInformerOption {
	return func(factory *multiNamespaceInformerFactory) *multiNamespaceInformerFactory {
		factory.scheme = scheme
		return factory
	}
}

// WithNamespaces sets the namespaces for a SharedInformerFactory.
func WithNamespaces(namespaces []string) SharedInformerOption {
	return func(factory *multiNamespaceInformerFactory) *multiNamespaceInformerFactory {
		factory.SetNamespaces(namespaces)
		return factory
	}
}

// WithTweakListOptions sets list options for a SharedInformerFactory.
func WithTweakListOptions(tweakListOptions dynamicinformer.TweakListOptionsFunc) SharedInformerOption {
	return func(factory *multiNamespaceInformerFactory) *multiNamespaceInformerFactory {
		factory.tweakListOptions = tweakListOptions
		return factory
	}
}

// WithCustomResyncConfig sets custom resync period for certain resources.
func WithCustomResyncConfig(config map[schema.GroupVersionResource]time.Duration) SharedInformerOption {
	return func(factory *multiNamespaceInformerFactory) *multiNamespaceInformerFactory {
		for gvr, resyncPeriod := range config {
			factory.customResync[gvr] = resyncPeriod
		}
		return factory
	}
}

// multiNamespaceInformerFactory provides a dynamic informer factory that
// creates informers which track changes across a set of namespaces.
type multiNamespaceInformerFactory struct {
	client           dynamic.Interface
	scheme           *runtime.Scheme
	resyncPeriod     time.Duration
	lock             sync.Mutex
	namespaces       sets.Set
	tweakListOptions dynamicinformer.TweakListOptionsFunc
	customResync     map[schema.GroupVersionResource]time.Duration

	// Map of created informers by resource type.
	informers map[schema.GroupVersionResource]*multiNamespaceGenericInformer
}

var _ SharedInformerFactory = &multiNamespaceInformerFactory{}

// NewSharedInformerFactory returns a new cross-namespace shared informer
// factory.  Use SetNamespaces on the resulting factory to configure the set of
// namespaces to be watched.
func NewSharedInformerFactory(client dynamic.Interface, resync time.Duration) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, resync)
}

// NewSharedInformerFactoryWithOptions constructs a new cross-namespace shared
// informer factory with the given options applied.  You must either supply the
// WithNamespaces option, or call SetNamespaces on the returned factory to
// configure the set of namespaces to be watched.
func NewSharedInformerFactoryWithOptions(client dynamic.Interface, resync time.Duration, options ...SharedInformerOption) SharedInformerFactory {
	factory := &multiNamespaceInformerFactory{
		client:       client,
		scheme:       scheme.Scheme,
		resyncPeriod: resync,
		informers:    make(map[schema.GroupVersionResource]*multiNamespaceGenericInformer),
		customResync: make(map[schema.GroupVersionResource]time.Duration),
	}

	for _, opt := range options {
		factory = opt(factory)
	}

	return factory
}

// GetScheme returns the runtime.Scheme for the factory.
func (f *multiNamespaceInformerFactory) GetScheme() *runtime.Scheme {
	return f.scheme
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
func (f *multiNamespaceInformerFactory) ForResource(gvr schema.GroupVersionResource, namespaced bool) informers.GenericInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	// Return existing informer if found.
	if informer, ok := f.informers[gvr]; ok {
		return informer
	}

	// Check for a custom resync period for this resource type.
	resyncPeriod, ok := f.customResync[gvr]
	if !ok {
		resyncPeriod = f.resyncPeriod
	}

	newInformerFunc := func(namespace string) informers.GenericInformer {
		// Namespace argument is ignored for cluster-scoped resources.
		// TODO: Should we use the discovery API to determine resource scope?
		if !namespaced {
			namespace = metav1.NamespaceAll
		}

		return dynamicinformer.NewFilteredDynamicInformer(
			f.client,
			gvr,
			namespace,
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			f.tweakListOptions,
		)
	}

	informer := NewMultiNamespaceInformer(namespaced, f.resyncPeriod, newInformerFunc)
	lister := dynamiclister.New(informer.GetIndexer(), gvr)

	for namespace := range f.namespaces {
		informer.AddNamespace(namespace)
	}

	f.informers[gvr] = &multiNamespaceGenericInformer{
		informer: informer,
		lister:   dynamiclister.NewRuntimeObjectShim(lister),
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
