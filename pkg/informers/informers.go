package informers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

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
	ForResource(resource schema.GroupVersionResource) informers.GenericInformer
	WaitForCacheSync(stopCh <-chan struct{})
}

// multiNamespaceInformerFactory provides a dynamic informer factory that
// creates informers which track changes across set of namespaces.
type multiNamespaceInformerFactory struct {
	namespaceToFactory map[string]dynamicinformer.DynamicSharedInformerFactory
}

var _ InformerFactory = &multiNamespaceInformerFactory{}

func NewInformerFactory(client dynamic.Interface, namespaces []string) (InformerFactory, error) {
	// TOOD: Dedupe namespaces and check for metav1.NamespaceAll
	if len(namespaces) < 1 {
		return nil, errors.New("must provide at least one namespace, which may be metav1.NamespaceAll")
	}

	resyncTime := 1 * time.Minute // TODO: This should be an arg.

	factory := &multiNamespaceInformerFactory{
		namespaceToFactory: make(map[string]dynamicinformer.DynamicSharedInformerFactory),
	}

	for _, namespace := range namespaces {
		factory.namespaceToFactory[namespace] =
			dynamicinformer.NewFilteredDynamicSharedInformerFactory(
				client,
				resyncTime,
				namespace,
				nil,
			)
	}

	return factory, nil
}

func (f *multiNamespaceInformerFactory) ForResource(gvr schema.GroupVersionResource) informers.GenericInformer {
	genericInformer := &multiNamespaceGenericInformer{
		informer: &multiNamespaceInformer{
			namespaceToInformer: make(map[string]cache.SharedIndexInformer),
		},
		lister: &multiNamespaceLister{
			namespaceToLister: make(map[string]cache.GenericLister),
			resource:          gvr,
		},
	}

	for ns, factory := range f.namespaceToFactory {
		i := factory.ForResource(gvr)
		genericInformer.informer.namespaceToInformer[ns] = i.Informer()
		genericInformer.lister.namespaceToLister[ns] = i.Lister()
	}

	return genericInformer
}

func (f *multiNamespaceInformerFactory) Start(stopCh <-chan struct{}) {
	for _, factory := range f.namespaceToFactory {
		factory.Start(stopCh)
	}
}

func (f *multiNamespaceInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) {
	for _, f := range f.namespaceToFactory {
		f.WaitForCacheSync(stopCh)
	}
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

// multiNamespaceInformer satisfies the SharedIndexInformer interface and
// provides an informer that works across a set of namespaces -- though not all
// methods are actually usable.
//
// TODO: Provide scoped down interfaces as well.
type multiNamespaceInformer struct {
	namespaceToInformer map[string]cache.SharedIndexInformer
}

var _ cache.SharedIndexInformer = &multiNamespaceInformer{}

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

func (i *multiNamespaceInformer) Run(stopCh <-chan struct{}) {
	panic("not implemented") // TODO: This could probably be implemented.
}

// AddEventHandler adds the handler to each namespaced informer.
func (i *multiNamespaceInformer) AddEventHandler(handler cache.ResourceEventHandler) {
	for _, informer := range i.namespaceToInformer {
		informer.AddEventHandler(handler)
	}
}

// AddEventHandlerWithResyncPeriod adds the handler with a resync period to each
// namespaced informer.
func (i *multiNamespaceInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	for _, informer := range i.namespaceToInformer {
		informer.AddEventHandlerWithResyncPeriod(handler, resyncPeriod)
	}
}

// AddIndexers adds the indexer for each namespaced informer.
func (i *multiNamespaceInformer) AddIndexers(indexers cache.Indexers) error {
	for _, informer := range i.namespaceToInformer {
		err := informer.AddIndexers(indexers)
		if err != nil {
			return err
		}
	}
	return nil
}

// HasSynced checks if each namespaced informer has synced.
func (i *multiNamespaceInformer) HasSynced() bool {
	for _, informer := range i.namespaceToInformer {
		if ok := informer.HasSynced(); !ok {
			return false
		}
	}
	return true
}

// multiNamespaceLister satisfies the GenericLister interface and works across a
// set of namespaces.
type multiNamespaceLister struct {
	namespaceToLister map[string]cache.GenericLister
	resource          schema.GroupVersionResource
}

var _ cache.GenericLister = &multiNamespaceLister{}

func (l *multiNamespaceLister) List(selector labels.Selector) ([]runtime.Object, error) {
	var result []runtime.Object

	for _, lister := range l.namespaceToLister {
		objs, err := lister.List(selector)
		if err != nil {
			return nil, err
		}

		result = append(result, objs...)
	}

	return result, nil
}

func (l *multiNamespaceLister) Get(name string) (runtime.Object, error) {
	namespace, _, err := cache.SplitMetaNamespaceKey(name)
	if err != nil {
		return nil, err
	}

	return l.ByNamespace(namespace).Get(name)
}

func (l *multiNamespaceLister) ByNamespace(namespace string) cache.GenericNamespaceLister {
	if lister, ok := l.namespaceToLister[metav1.NamespaceAll]; ok {
		return lister.ByNamespace(namespace)
	} else if lister, ok := l.namespaceToLister[namespace]; ok {
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
