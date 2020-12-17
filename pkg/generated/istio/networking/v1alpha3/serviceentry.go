// Code generated by xns-informer-gen. DO NOT EDIT.

package v1alpha3

import (
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	informers "istio.io/client-go/pkg/informers/externalversions/networking/v1alpha3"
	listers "istio.io/client-go/pkg/listers/networking/v1alpha3"
	"k8s.io/client-go/tools/cache"
)

type serviceEntryInformer struct {
	informer cache.SharedIndexInformer
}

var _ informers.ServiceEntryInformer = &serviceEntryInformer{}

func NewServiceEntryInformer(f xnsinformers.SharedInformerFactory) informers.ServiceEntryInformer {
	resource := v1alpha3.SchemeGroupVersion.WithResource("serviceentries")
	converter := xnsinformers.NewListWatchConverter(
		f.GetScheme(),
		&v1alpha3.ServiceEntry{},
		&v1alpha3.ServiceEntryList{},
	)

	informer := f.ForResource(resource, xnsinformers.ResourceOptions{
		ClusterScoped:      false,
		ListWatchConverter: converter,
	})

	return &serviceEntryInformer{informer: informer.Informer()}
}

func (i *serviceEntryInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *serviceEntryInformer) Lister() listers.ServiceEntryLister {
	return listers.NewServiceEntryLister(i.informer.GetIndexer())
}
