package informers

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type cacheConverter struct {
	indexer   cache.Indexer
	outType   reflect.Type
	converter runtime.UnstructuredConverter
}

var _ cache.Indexer = &cacheConverter{}

// NewCacheConverter returns a new cache converter that wraps the given
// cache.Indexer, which is expected to hold unstructured objects.  Those objects
// should represent the type backing the given runtime.Object.  Items read from
// the cache will be converted to the concrete type, and items written to the
// cache will be converted to an unstructured object.
func NewCacheConverter(idx cache.Indexer, obj runtime.Object) *cacheConverter {
	return &cacheConverter{
		indexer:   idx,
		outType:   reflect.ValueOf(obj).Elem().Type(),
		converter: runtime.DefaultUnstructuredConverter,
	}
}

func (c *cacheConverter) Add(obj interface{}) error {
	unstructuredObj, err := c.converter.ToUnstructured(obj)
	if err != nil {
		return nil
	}

	return c.indexer.Add(&unstructured.Unstructured{Object: unstructuredObj})
}

func (c *cacheConverter) Update(obj interface{}) error {
	unstructuredObj, err := c.converter.ToUnstructured(obj)
	if err != nil {
		return nil
	}

	return c.indexer.Update(&unstructured.Unstructured{Object: unstructuredObj})
}

func (c *cacheConverter) Delete(obj interface{}) error {
	unstructuredObj, err := c.converter.ToUnstructured(obj)
	if err != nil {
		return nil
	}

	return c.indexer.Delete(&unstructured.Unstructured{Object: unstructuredObj})
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
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expected *unstructured.Unstructured, got %T", obj)
	}

	out := reflect.New(c.outType).Interface()

	if err := c.converter.FromUnstructured(u.Object, out); err != nil {
		return nil, err
	}

	return out, nil
}
