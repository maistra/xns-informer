// Code generated by xns-informer-gen. DO NOT EDIT.

package v1alpha1

import (
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	"k8s.io/api/flowcontrol/v1alpha1"
	informers "k8s.io/client-go/informers/flowcontrol/v1alpha1"
	listers "k8s.io/client-go/listers/flowcontrol/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type priorityLevelConfigurationInformer struct {
	informer cache.SharedIndexInformer
}

var _ informers.PriorityLevelConfigurationInformer = &priorityLevelConfigurationInformer{}

func NewPriorityLevelConfigurationInformer(f xnsinformers.SharedInformerFactory) informers.PriorityLevelConfigurationInformer {
	resource := v1alpha1.SchemeGroupVersion.WithResource("prioritylevelconfigurations")
	informer := f.ClusterResource(resource).Informer()

	return &priorityLevelConfigurationInformer{
		informer: xnsinformers.NewInformerConverter(f.GetScheme(), informer, &v1alpha1.PriorityLevelConfiguration{}),
	}
}

func (i *priorityLevelConfigurationInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *priorityLevelConfigurationInformer) Lister() listers.PriorityLevelConfigurationLister {
	return listers.NewPriorityLevelConfigurationLister(i.informer.GetIndexer())
}
