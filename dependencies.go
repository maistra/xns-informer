package xnsinformer

import (
	// import openshift route API
	_ "github.com/openshift/api/route/v1"
	// import gateway API so we can generate the informer factory
	_ "sigs.k8s.io/gateway-api/conformance/utils/suite"

	// import istio API to generate the istio informer factory
	_ "istio.io/client-go/pkg/clientset/versioned"
)
