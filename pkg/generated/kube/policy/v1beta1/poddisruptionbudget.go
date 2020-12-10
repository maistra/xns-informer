// Code generated by xns-informer-gen. DO NOT EDIT.

package v1beta1

import (
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	v1beta1 "k8s.io/api/policy/v1beta1"
	informers "k8s.io/client-go/informers/policy/v1beta1"
	listers "k8s.io/client-go/listers/policy/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type podDisruptionBudgetInformer struct {
	informer cache.SharedIndexInformer
}

var _ informers.PodDisruptionBudgetInformer = &podDisruptionBudgetInformer{}

func NewPodDisruptionBudgetInformer(f xnsinformers.SharedInformerFactory) informers.PodDisruptionBudgetInformer {
	resource := v1beta1.SchemeGroupVersion.WithResource("poddisruptionbudgets")
	converter := xnsinformers.NewListWatchConverter(
		f.GetScheme(),
		&v1beta1.PodDisruptionBudget{},
		&v1beta1.PodDisruptionBudgetList{},
	)

	informer := f.ForResource(resource, xnsinformers.ResourceOptions{
		ClusterScoped:      false,
		ListWatchConverter: converter,
	})

	return &podDisruptionBudgetInformer{informer: informer.Informer()}
}

func (i *podDisruptionBudgetInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *podDisruptionBudgetInformer) Lister() listers.PodDisruptionBudgetLister {
	return listers.NewPodDisruptionBudgetLister(i.informer.GetIndexer())
}
