// Code generated by xns-informer-gen. DO NOT EDIT.

package v1

import (
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	v1 "k8s.io/api/networking/v1"
	informers "k8s.io/client-go/informers/networking/v1"
	listers "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
)

type ingressClassInformer struct {
	informer cache.SharedIndexInformer
}

var _ informers.IngressClassInformer = &ingressClassInformer{}

func NewIngressClassInformer(f xnsinformers.SharedInformerFactory) informers.IngressClassInformer {
	resource := v1.SchemeGroupVersion.WithResource("ingressclasses")
	converter := xnsinformers.NewListWatchConverter(
		f.GetScheme(),
		&v1.IngressClass{},
		&v1.IngressClassList{},
	)

	informer := f.ForResource(resource, xnsinformers.ResourceOptions{
		ClusterScoped:      true,
		ListWatchConverter: converter,
	})

	return &ingressClassInformer{informer: informer.Informer()}
}

func (i *ingressClassInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *ingressClassInformer) Lister() listers.IngressClassLister {
	return listers.NewIngressClassLister(i.informer.GetIndexer())
}
