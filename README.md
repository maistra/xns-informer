# Cross Namespace Informers

[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/maistra/xns-informer)](https://pkg.go.dev/mod/github.com/maistra/xns-informer)
[![Go Report Card](https://goreportcard.com/badge/github.com/maistra/xns-informer)](https://goreportcard.com/report/github.com/maistra/xns-informer)
[![CI](https://github.com/maistra/xns-informer/workflows/CI/badge.svg)](https://github.com/maistra/xns-informer/actions)

Kubernetes informers across a dynamic set of namespaces.

**Status:** This is experimental. Don't expect API stability yet.


## Purpose

The Kubernetes client libraries provide informer objects, which allow clients
to watch and react to changes in a set of resources.  Unfortunately, they only
work either for a single namespace or for all namespaces.  In some cases, a
client may not haves permission to list and read objects across the entire
cluster, but may still want to watch more than one namespace.

This library provides a method to create informers for any resource type that
work across a dynamic set of namespaces, as well as a code generation tool that
can generate packages containing informer factories for sets of types that are
API compatible with existing versions.  A cross-namespace version of the
informer factory for Kubernetes API types is included.


## Example

An example of creating an informer for ConfigMap resources:

```golang
import kubeinformers "github.com/maistra/xns-informer/pkg/generated/kube"

// Create a new shared informer factory.
// Assume client is a Kubernetes client interface.
factory := kubeinformers.NewSharedInformerFactoryWithOptions(
	client,
	1*time.Minute,
	kubeinformers.WithNamespaces("default", "kube-system"),
)

// Create an informer for ConfigMap resources.
informer := factory.Core().V1().ConfigMaps()

// Add an event handler to the new informer.
informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	AddFunc: func(obj interface{}) {
		log.Print("ConfigMap add event!")
	},
})

stopCh := make(chan struct{})

// Start all informers and wait for their caches to sync.
factory.Start(stopCh)
factory.WaitForCacheSync(stopCh)

// You can list all ConfigMap objects across only the tracked namespaces.
list, err := informer.Lister().List(labels.Everything())

// You can change the set of namespaces at anytime.
factory.SetNamespaces("my-application", "another-application")

// This will now only list ConfigMaps in the new namespaces.
list, err = informer.Lister().List(labels.Everything())
```

See the [examples][1] directory for more detailed examples.


## Code Generation

This library includes a code generation tool, `xns-informer-gen`, based on the
Kubernetes `informer-gen` tool.  It can generate packages containing informer
factories that return interfaces which are API compatible with those generated
by `informer-gen`.  For an example, See the [update-codegen.sh][2] script, and
the [package][3] it generates.


  [1]: https://github.com/maistra/xns-informer/tree/master/examples
  [2]: https://github.com/maistra/xns-informer/blob/master/hack/update-codegen.sh
  [3]: https://github.com/maistra/xns-informer/tree/master/pkg/generated/kube
