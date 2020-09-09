package main

import (
	"os"

	"github.com/spf13/pflag"
	"k8s.io/gengo/args"
	"k8s.io/klog/v2"

	"github.com/maistra/xns-informer/cmd/xns-informer-gen/generators"
)

func main() {
	klog.InitFlags(nil)

	arguments := args.Default()
	customArgs := &generators.CustomArgs{
		ListersPackage:   "k8s.io/client-go/listers",
		InformersPackage: "k8s.io/client-go/informers",
	}

	pflag.CommandLine.StringVar(&customArgs.ListersPackage, "listers-package",
		customArgs.ListersPackage, "Base import path for listers")
	pflag.CommandLine.StringVar(&customArgs.InformersPackage, "informers-package",
		customArgs.InformersPackage, "Base import path for informers")

	arguments.CustomArgs = customArgs

	if err := arguments.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	); err != nil {
		klog.Errorf("Error: %v", err)
		os.Exit(1)
	}
	klog.V(2).Info("Completed successfully.")
}
