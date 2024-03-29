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

package v1alpha2

import (
	informers "github.com/maistra/xns-informer/pkg/informers"
	v1alpha2 "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions/apis/v1alpha2"
	internalinterfaces "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions/internalinterfaces"
)

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespaces       informers.NamespaceSet
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespaces informers.NamespaceSet, tweakListOptions internalinterfaces.TweakListOptionsFunc) v1alpha2.Interface {
	return &version{factory: f, namespaces: namespaces, tweakListOptions: tweakListOptions}
}

// GRPCRoutes returns a GRPCRouteInformer.
func (v *version) GRPCRoutes() v1alpha2.GRPCRouteInformer {
	return &gRPCRouteInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// Gateways returns a GatewayInformer.
func (v *version) Gateways() v1alpha2.GatewayInformer {
	return &gatewayInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// GatewayClasses returns a GatewayClassInformer.
func (v *version) GatewayClasses() v1alpha2.GatewayClassInformer {
	return &gatewayClassInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// HTTPRoutes returns a HTTPRouteInformer.
func (v *version) HTTPRoutes() v1alpha2.HTTPRouteInformer {
	return &hTTPRouteInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// ReferenceGrants returns a ReferenceGrantInformer.
func (v *version) ReferenceGrants() v1alpha2.ReferenceGrantInformer {
	return &referenceGrantInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// TCPRoutes returns a TCPRouteInformer.
func (v *version) TCPRoutes() v1alpha2.TCPRouteInformer {
	return &tCPRouteInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// TLSRoutes returns a TLSRouteInformer.
func (v *version) TLSRoutes() v1alpha2.TLSRouteInformer {
	return &tLSRouteInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}

// UDPRoutes returns a UDPRouteInformer.
func (v *version) UDPRoutes() v1alpha2.UDPRouteInformer {
	return &uDPRouteInformer{factory: v.factory, namespaces: v.namespaces, tweakListOptions: v.tweakListOptions}
}
