package informers

import (
	"errors"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/cache"
)

var ErrCacheReadOnly = errors.New("cache is read-only")

type cacheReader struct {
	informer *multiNamespaceInformer
}

var _ cache.Indexer = &cacheReader{}

// NewCacheReader returns a new cache reader which satisfies the cache.Indexer
// interface and provides read-only access across the underlying caches for the
// given cross-namespace informer.
func NewCacheReader(informer *multiNamespaceInformer) *cacheReader {
	return &cacheReader{informer: informer}
}

func (c *cacheReader) Add(obj interface{}) error {
	return ErrCacheReadOnly
}

func (c *cacheReader) Update(obj interface{}) error {
	return ErrCacheReadOnly
}

func (c *cacheReader) Delete(obj interface{}) error {
	return ErrCacheReadOnly
}

func (c *cacheReader) List() (res []interface{}) {
	for _, idx := range c.informer.getIndexers() {
		res = append(res, idx.List()...)
	}

	return res
}

func (c *cacheReader) ListKeys() (res []string) {
	for _, idx := range c.informer.getIndexers() {
		res = append(res, idx.ListKeys()...)
	}

	return res
}

func (c *cacheReader) GetIndexers() cache.Indexers {
	res := make(cache.Indexers)
	for _, idx := range c.informer.getIndexers() {
		for k, v := range idx.GetIndexers() {
			res[k] = v
		}
	}

	return res
}

func (c *cacheReader) Index(indexName string, obj interface{}) (res []interface{}, err error) {
	for _, idx := range c.informer.getIndexers() {
		objs, err := idx.Index(indexName, obj)
		if err != nil {
			return nil, err
		}

		res = append(res, objs...)
	}

	return res, nil
}

func (c *cacheReader) IndexKeys(indexName, indexKey string) (res []string, err error) {
	for _, idx := range c.informer.getIndexers() {
		keys, err := idx.IndexKeys(indexName, indexKey)
		if err != nil {
			return nil, err
		}

		res = append(res, keys...)
	}

	return res, nil
}

func (c *cacheReader) ListIndexFuncValues(indexName string) (res []string) {
	for _, idx := range c.informer.getIndexers() {
		res = append(res, idx.ListIndexFuncValues(indexName)...)
	}

	return res
}

func (c *cacheReader) ByIndex(indexName, indexKey string) (res []interface{}, err error) {
	for _, idx := range c.informer.getIndexers() {
		keys, err := idx.ByIndex(indexName, indexKey)
		if err != nil {
			return nil, err // TODO: Multi-error?
		}

		res = append(res, keys...)
	}

	return res, nil
}

func (c *cacheReader) AddIndexers(newIndexers cache.Indexers) error {
	return ErrCacheReadOnly
}

func (c *cacheReader) Get(obj interface{}) (item interface{}, exists bool, err error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, false, err
	}

	namespace := accessor.GetNamespace()
	indexers := c.informer.getIndexers()

	if _, ok := indexers[namespace]; !ok {
		return nil, false, nil // TODO: Should this return an error?
	}

	return indexers[namespace].Get(obj)
}

func (c *cacheReader) GetByKey(key string) (item interface{}, exists bool, err error) {
	namespace, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, false, err
	}

	indexers := c.informer.getIndexers()
	if _, ok := indexers[namespace]; !ok {
		return nil, false, nil // TODO: Should this return an error?
	}

	return indexers[namespace].GetByKey(key)
}

func (c *cacheReader) Replace(list []interface{}, resourceVersion string) error {
	return ErrCacheReadOnly
}

func (c *cacheReader) Resync() error {
	return nil
}
