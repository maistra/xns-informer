package informers

import (
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// cacheConverter wraps caches and converts objects types.
type cacheConverter struct {
	indexer   cache.Indexer
	outType   reflect.Type
	converter runtime.ObjectConvertor
}

var _ cache.Indexer = &cacheConverter{}

// NewCacheConverter returns a new cache converter that wraps the given
// cache.Indexer, which is expected to hold unstructured objects.  Those objects
// should represent the type backing the given runtime.Object.  Items read from
// the cache will be converted to the concrete type, and items written to the
// cache will be converted to an unstructured object.
func NewCacheConverter(conv runtime.ObjectConvertor, idx cache.Indexer, obj runtime.Object) *cacheConverter {
	return &cacheConverter{
		indexer:   idx,
		outType:   reflect.ValueOf(obj).Elem().Type(),
		converter: conv,
	}
}

func (c *cacheConverter) Add(obj interface{}) error {
	u := &unstructured.Unstructured{}
	if err := c.converter.Convert(obj, u, nil); err != nil {
		return err
	}

	return c.indexer.Add(u)
}

func (c *cacheConverter) Update(obj interface{}) error {
	u := &unstructured.Unstructured{}
	if err := c.converter.Convert(obj, u, nil); err != nil {
		return err
	}

	return c.indexer.Update(u)
}

func (c *cacheConverter) Delete(obj interface{}) error {
	u := &unstructured.Unstructured{}
	if err := c.converter.Convert(obj, u, nil); err != nil {
		return err
	}

	return c.indexer.Delete(u)
}

func (c *cacheConverter) List() (res []interface{}) {
	var err error

	res, err = c.convertList(c.indexer.List())
	if err != nil {
		klog.Errorf("cache converter list error: %v", err)
	}

	return res
}

func (c *cacheConverter) ListKeys() (res []string) {
	return c.indexer.ListKeys()
}

func (c *cacheConverter) GetIndexers() cache.Indexers {
	return c.indexer.GetIndexers()
}

func (c *cacheConverter) Index(indexName string, obj interface{}) (res []interface{}, err error) {
	objs, err := c.indexer.Index(indexName, obj)
	if err != nil {
		return nil, err
	}

	return c.convertList(objs)
}

func (c *cacheConverter) IndexKeys(indexName, indexKey string) (res []string, err error) {
	return c.indexer.IndexKeys(indexName, indexKey)
}

func (c *cacheConverter) ListIndexFuncValues(indexName string) (res []string) {
	return c.indexer.ListIndexFuncValues(indexName)
}

func (c *cacheConverter) ByIndex(indexName, indexKey string) (res []interface{}, err error) {
	objs, err := c.indexer.ByIndex(indexName, indexKey)
	if err != nil {
		return nil, err
	}

	return c.convertList(objs)
}

func (c *cacheConverter) AddIndexers(newIndexers cache.Indexers) error {
	return c.indexer.AddIndexers(newIndexers)
}

func (c *cacheConverter) Get(obj interface{}) (item interface{}, exists bool, err error) {
	item, exists, err = c.indexer.Get(obj)
	if err != nil || !exists {
		return nil, exists, err
	}

	item, err = c.convert(item)
	return item, exists, err
}

func (c *cacheConverter) GetByKey(key string) (item interface{}, exists bool, err error) {
	item, exists, err = c.indexer.GetByKey(key)
	if err != nil || !exists {
		return nil, exists, err
	}

	item, err = c.convert(item)
	return item, exists, err
}

func (c *cacheConverter) Replace(list []interface{}, resourceVersion string) error {
	return c.indexer.Replace(list, resourceVersion)
}

func (c *cacheConverter) Resync() error {
	return c.indexer.Resync()
}

func (c *cacheConverter) convertList(objs []interface{}) (res []interface{}, err error) {
	for _, obj := range objs {
		out, err := c.convert(obj)
		if err != nil {
			return nil, err
		}

		res = append(res, out)
	}

	return res, nil
}

func (c *cacheConverter) convert(obj interface{}) (res interface{}, err error) {
	res = reflect.New(c.outType).Interface()
	err = c.converter.Convert(obj, res, nil)

	return res, err
}

// informerConverter wraps informers and converts objects types.
type informerConverter struct {
	informer  cache.SharedIndexInformer
	indexer   cache.Indexer
	outType   reflect.Type
	converter runtime.ObjectConvertor
}

var _ cache.SharedIndexInformer = &informerConverter{}

// NewInformerConverter returns a new informer converter that wraps the given
// cache.SharedIndexInformer.  It wraps any event handlers added to the informer
// and attempts to convert the objects passed to them into the configured
// concrete type.  It also creates a new cache converter that will be returned
// when GetStore or GetIndexer is called.
func NewInformerConverter(conv runtime.ObjectConvertor, informer cache.SharedIndexInformer, obj runtime.Object) *informerConverter {
	return &informerConverter{
		informer:  informer,
		indexer:   NewCacheConverter(conv, informer.GetIndexer(), obj),
		outType:   reflect.ValueOf(obj).Elem().Type(),
		converter: conv,
	}
}

func (i *informerConverter) GetController() cache.Controller {
	return i.informer.GetController()
}

func (i *informerConverter) GetStore() cache.Store {
	return i.indexer
}

func (i *informerConverter) GetIndexer() cache.Indexer {
	return i.indexer
}

func (i *informerConverter) LastSyncResourceVersion() string {
	return i.informer.LastSyncResourceVersion()
}

func (i *informerConverter) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	return i.informer.SetWatchErrorHandler(handler)
}

func (i *informerConverter) Run(stopCh <-chan struct{}) {
	i.informer.Run(stopCh)
}

func (i *informerConverter) AddEventHandler(handler cache.ResourceEventHandler) {
	wrapped := i.wrapHandler(handler)
	i.informer.AddEventHandler(wrapped)
}

func (i *informerConverter) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	wrapped := i.wrapHandler(handler)
	i.informer.AddEventHandlerWithResyncPeriod(wrapped, resyncPeriod)
}

func (i *informerConverter) AddIndexers(indexers cache.Indexers) error {
	return i.informer.AddIndexers(indexers)
}

func (i *informerConverter) HasSynced() bool {
	return i.informer.HasSynced()
}

// wrapHandler wraps the given cache.ResourceEventHandler and attempts to
// convert the arguments passed to each handler function to the converter's
// concrete type.  If the objects can't be converted, the handlers are called
// with the original unconverted objects.
func (i *informerConverter) wrapHandler(handler cache.ResourceEventHandler) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			obj = i.mayConvert(obj)
			handler.OnAdd(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldObj = i.mayConvert(oldObj)
			newObj = i.mayConvert(newObj)
			handler.OnUpdate(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			obj = i.mayConvert(obj)
			handler.OnDelete(obj)
		},
	}
}

// mayConvert attempts to convert the given object into the converter's concrete
// type, but ignores failures, returning the unconverted object as-is.
func (i *informerConverter) mayConvert(obj interface{}) interface{} {
	out := reflect.New(i.outType).Interface()
	if err := i.converter.Convert(obj, out, nil); err != nil {
		return obj
	}

	return out
}
