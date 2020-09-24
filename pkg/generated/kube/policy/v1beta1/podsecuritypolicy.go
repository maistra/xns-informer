// Code generated by xns-informer-gen. DO NOT EDIT.

package v1beta1

import (
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	"k8s.io/api/policy/v1beta1"
	informers "k8s.io/client-go/informers/policy/v1beta1"
	listers "k8s.io/client-go/listers/policy/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type podSecurityPolicyInformer struct {
	informer cache.SharedIndexInformer
}

var _ informers.PodSecurityPolicyInformer = &podSecurityPolicyInformer{}

func NewPodSecurityPolicyInformer(f xnsinformers.SharedInformerFactory) informers.PodSecurityPolicyInformer {
	resource := v1beta1.SchemeGroupVersion.WithResource("podsecuritypolicies")
	informer := f.ClusterResource(resource).Informer()

	return &podSecurityPolicyInformer{
		informer: xnsinformers.NewInformerConverter(f.GetScheme(), informer, &v1beta1.PodSecurityPolicy{}),
	}
}

func (i *podSecurityPolicyInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *podSecurityPolicyInformer) Lister() listers.PodSecurityPolicyLister {
	return listers.NewPodSecurityPolicyLister(i.informer.GetIndexer())
}
