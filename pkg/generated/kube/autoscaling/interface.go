/*
Copyright Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by xns-informer-gen. DO NOT EDIT.

package autoscaling

import (
	autoscalingv1 "github.com/maistra/xns-informer/pkg/generated/kube/autoscaling/v1"
	autoscalingv2 "github.com/maistra/xns-informer/pkg/generated/kube/autoscaling/v2"
	autoscalingv2beta1 "github.com/maistra/xns-informer/pkg/generated/kube/autoscaling/v2beta1"
	autoscalingv2beta2 "github.com/maistra/xns-informer/pkg/generated/kube/autoscaling/v2beta2"
	informers "github.com/maistra/xns-informer/pkg/informers"
	autoscaling "k8s.io/client-go/informers/autoscaling"
	v1 "k8s.io/client-go/informers/autoscaling/v1"
	v2 "k8s.io/client-go/informers/autoscaling/v2"
	v2beta1 "k8s.io/client-go/informers/autoscaling/v2beta1"
	v2beta2 "k8s.io/client-go/informers/autoscaling/v2beta2"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
)

type group struct {
	factory          internalinterfaces.SharedInformerFactory
	namespaces       informers.NamespaceSet
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespaces informers.NamespaceSet, tweakListOptions internalinterfaces.TweakListOptionsFunc) autoscaling.Interface {
	return &group{factory: f, namespaces: namespaces, tweakListOptions: tweakListOptions}
}

// V1 returns a new v1.Interface.
func (g *group) V1() v1.Interface {
	return autoscalingv1.New(g.factory, g.namespaces, g.tweakListOptions)
}

// V2 returns a new v2.Interface.
func (g *group) V2() v2.Interface {
	return autoscalingv2.New(g.factory, g.namespaces, g.tweakListOptions)
}

// V2beta1 returns a new v2beta1.Interface.
func (g *group) V2beta1() v2beta1.Interface {
	return autoscalingv2beta1.New(g.factory, g.namespaces, g.tweakListOptions)
}

// V2beta2 returns a new v2beta2.Interface.
func (g *group) V2beta2() v2beta2.Interface {
	return autoscalingv2beta2.New(g.factory, g.namespaces, g.tweakListOptions)
}
