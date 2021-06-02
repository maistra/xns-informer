package main

import (
	"flag"
	"os"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	kubeinformers "github.com/maistra/xns-informer/pkg/generated/kube"
)

// This is an example of using cross-namespace informers for Kubernetes types.
//
// Set up your Kubernetes client config, and run this passing a set of
// namespaces, e.g.: go run main.go default ns-two ns-three

func main() {
	klog.InitFlags(nil)
	flag.Set("v", "4")
	flag.Parse()

	configOverrides := &clientcmd.ConfigOverrides{}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		klog.Fatal(err)
	}

	client := kubernetes.NewForConfigOrDie(config)
	resync := 1 * time.Minute
	stopCh := make(chan struct{})
	namespaces := os.Args[1:]

	klog.Infof("Creating informer factory for namespaces: %v", namespaces)

	// Create the factory for Kubernetes informers.
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(
		client,
		resync,
		kubeinformers.WithNamespaces(namespaces...),
	)

	// Get a cross-namespace informer for ConfigMaps.
	cmInformer := kubeInformerFactory.Core().V1().ConfigMaps()

	// Add handlers that just log the events.
	cmInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			m, err := apimeta.Accessor(obj)
			if err != nil {
				klog.Fatal(err)
			}
			klog.Infof("*** Add Event for ConfigMap: %s/%s", m.GetNamespace(), m.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			m, err := apimeta.Accessor(newObj)
			if err != nil {
				klog.Fatal(err)
			}
			klog.Infof("*** Update Event for ConfigMap: %s/%s", m.GetNamespace(), m.GetName())
		},
		DeleteFunc: func(obj interface{}) {
			m, err := apimeta.Accessor(obj)
			if err != nil {
				klog.Fatal(err)
			}
			klog.Infof("*** Delete Event for ConfigMap: %s/%s", m.GetNamespace(), m.GetName())
		},
	})

	// After requesting new informers Start() and WaitForCacheSync() must be
	// called on the factory.  They can safely be called multiple times.
	kubeInformerFactory.Start(stopCh)
	kubeInformerFactory.WaitForCacheSync(stopCh)
	klog.Info("Informers started and synced!")

	firstNS := namespaces[0]

	klog.Infof("Listing all ConfigMaps in %q namespace...", firstNS)
	// Get a lister for the first tracked namespace, and list all ConfigMaps there.
	configMaps, err := cmInformer.Lister().ConfigMaps(firstNS).List(labels.Everything())
	if err != nil {
		klog.Fatal(err)
	}

	for i := range configMaps {
		namespace := configMaps[i].GetNamespace()
		name := configMaps[i].GetName()
		klog.Infof("  - %s/%s", namespace, name)
	}

	klog.Info("Listing and fetching ConfigMaps in all tracked namespaces...")
	// Get a lister across all tracked namespaces, and get all ConfigMaps.
	configMaps, err = cmInformer.Lister().List(labels.Everything())
	if err != nil {
		klog.Fatal(err)
	}

	for i := range configMaps {
		namespace := configMaps[i].GetNamespace()
		name := configMaps[i].GetName()
		klog.Infof("  - %s/%s", namespace, name)

		cm, err := cmInformer.Lister().ConfigMaps(namespace).Get(name)
		if err != nil {
			klog.Fatal(err)
		}

		klog.Infof("    - DATA: %v", cm.Data)
	}

	klog.Info("Waiting for new events. Press Ctrl-c to exit.")

	select {}
}
