package kube

import (
	"github.com/maistra/xns-informer/pkg/informers"
	"github.com/maistra/xns-informer/pkg/informers/kube/core"
)

type kubeInformerFactory struct {
	factory informers.InformerFactory
}

func NewKubeInformerFactory(f informers.InformerFactory) *kubeInformerFactory {
	return &kubeInformerFactory{factory: f}
}

func (f *kubeInformerFactory) Core() core.Interface {
	return core.New(f.factory)
}
