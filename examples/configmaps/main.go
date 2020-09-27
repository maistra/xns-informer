package main

import (
	"log"
	"os"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	kubeinformers "github.com/maistra/xns-informer/pkg/generated/kube"
	xnsinfomers "github.com/maistra/xns-informer/pkg/informers"
)

func main() {
	configOverrides := &clientcmd.ConfigOverrides{}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	resync := 1 * time.Minute
	stopCh := make(chan struct{})
	namespaces := os.Args[1:]
	log.Printf("Creating informer for namespaces: %v", namespaces)

	factory := xnsinfomers.NewSharedInformerFactoryWithOptions(
		client,
		resync,
		xnsinfomers.WithNamespaces(namespaces),
	)

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(factory)
	cmInformer := kubeInformerFactory.Core().V1().ConfigMaps()

	cmInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			m, err := apimeta.Accessor(obj)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("*** Add Event for ConfigMap: %s/%s", m.GetNamespace(), m.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			m, err := apimeta.Accessor(newObj)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("*** Update Event for ConfigMap: %s/%s", m.GetNamespace(), m.GetName())
		},
		DeleteFunc: func(obj interface{}) {
			m, err := apimeta.Accessor(obj)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("*** Delete Event for ConfigMap: %s/%s", m.GetNamespace(), m.GetName())
		},
	})

	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)
	log.Print("Informers started and synced!")

	log.Print("Listing all ConfigMaps in default namespace...")
	configMaps, err := cmInformer.Lister().ConfigMaps("default").List(labels.Everything())
	if err != nil {
		log.Fatal(err)
	}

	for i := range configMaps {
		namespace := configMaps[i].GetNamespace()
		name := configMaps[i].GetName()
		log.Printf("  - %s/%s", namespace, name)
	}

	log.Print("Listing and fetching all ConfigMaps...")
	configMaps, err = cmInformer.Lister().List(labels.Everything())
	if err != nil {
		log.Fatal(err)
	}

	for i := range configMaps {
		namespace := configMaps[i].GetNamespace()
		name := configMaps[i].GetName()
		log.Printf("  - %s/%s", namespace, name)

		cm, err := cmInformer.Lister().ConfigMaps(namespace).Get(name)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("    - DATA: %v", cm.Data)
	}

	select {}
}
