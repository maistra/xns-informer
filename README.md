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
work across a dynamic set of namespaces, as well as a code generation tool
that can generate packages containing informer factories for sets of types
that are (mostly) API compatible with existing versions.  A cross-namespace
version of the informer factory for Kubernetes API types is included.


## Example

An example of creating an informer for ConfigMap resources:

```golang
// Create a new shared informer factory.
// Assume client is a dynamic.Interface Kubernetes client.
factory := xnsinformers.NewSharedInformerFactoryWithOptions(
	client,
	1*time.Minute,
	xnsinformers.WithNamespaces([]string{"default", "application"}),
)

// Create an informer for ConfigMap resources.
resource := corev1.SchemeGroupVersion.WithResource("configmaps")
informer := factory.NamespacedResource(resource)

// Add an event handler to the new informer.
informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	AddFunc: func(obj interface{}) {
		// obj is an *unstructured.Unstructured here.
		log.Print("ConfigMap add event!")
	},
})

stopCh := make(chan struct{})

// Start all informers and wait for their caches to sync.
factory.Start(stopCh)
factory.WaitForCacheSync(stopCh)

// Now you can list all ConfigMap objects across the namespaces.
// The list will contain a slice of unstructured.Unstructured objects.
list, err := informer.Lister().List(labels.Everything())

// Want to do the above, but work with the concrete types?
kubeInformerFactory := kubeinformers.NewSharedInformerFactory(factory)
cmInformer := kubeInformerFactory.Core().V1().ConfigMaps()

// These are safe to call multiple times.
factory.Start(stopCh)
factory.WaitForCacheSync(stopCh)

// The same list operation as above, but returns ConfigMap objects.
configMaps, err = cmInformer.Lister().List(labels.Everything())
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
