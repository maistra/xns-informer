// Code generated by xns-informer-gen. DO NOT EDIT.

package v1beta1

import (
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	"k8s.io/api/discovery/v1beta1"
	informers "k8s.io/client-go/informers/discovery/v1beta1"
	listers "k8s.io/client-go/listers/discovery/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type endpointSliceInformer struct {
	informer cache.SharedIndexInformer
}

var _ informers.EndpointSliceInformer = &endpointSliceInformer{}

func NewEndpointSliceInformer(f xnsinformers.SharedInformerFactory) informers.EndpointSliceInformer {
	resource := v1beta1.SchemeGroupVersion.WithResource("endpointslices")
	informer := f.NamespacedResource(resource).Informer()

	return &endpointSliceInformer{
		informer: xnsinformers.NewInformerConverter(f.GetScheme(), informer, &v1beta1.EndpointSlice{}),
	}
}

func (i *endpointSliceInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *endpointSliceInformer) Lister() listers.EndpointSliceLister {
	return listers.NewEndpointSliceLister(i.informer.GetIndexer())
}
