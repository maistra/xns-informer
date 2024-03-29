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

package v1beta1

import (
	informers "github.com/maistra/xns-informer/pkg/informers"
	internalinterfaces "istio.io/client-go/pkg/informers/externalversions/internalinterfaces"
	v1beta1 "istio.io/client-go/pkg/informers/externalversions/networking/v1beta1"
)

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespaces       informers.NamespaceSet
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespaces informers.NamespaceSet, tweakListOptions internalinterfaces.TweakListOptionsFunc) v1beta1.Interface {
	return &version{factory: f, namespaces: namespaces, tweakListOptions: tweakListOptions}
}

// DestinationRules returns a DestinationRuleInformer.
func (v *version) DestinationRules() v1beta1.DestinationRuleInformer {
	return &destinationRuleInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// Gateways returns a GatewayInformer.
func (v *version) Gateways() v1beta1.GatewayInformer {
	return &gatewayInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// ProxyConfigs returns a ProxyConfigInformer.
func (v *version) ProxyConfigs() v1beta1.ProxyConfigInformer {
	return &proxyConfigInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// ServiceEntries returns a ServiceEntryInformer.
func (v *version) ServiceEntries() v1beta1.ServiceEntryInformer {
	return &serviceEntryInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// Sidecars returns a SidecarInformer.
func (v *version) Sidecars() v1beta1.SidecarInformer {
	return &sidecarInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// VirtualServices returns a VirtualServiceInformer.
func (v *version) VirtualServices() v1beta1.VirtualServiceInformer {
	return &virtualServiceInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// WorkloadEntries returns a WorkloadEntryInformer.
func (v *version) WorkloadEntries() v1beta1.WorkloadEntryInformer {
	return &workloadEntryInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// WorkloadGroups returns a WorkloadGroupInformer.
func (v *version) WorkloadGroups() v1beta1.WorkloadGroupInformer {
	return &workloadGroupInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}
