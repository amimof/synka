package main

import (
	"github.com/amimof/synka/pkg/controller"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"os"
	"os/signal"
)

var (
	masterURL  string
	kubeconfig string
	informers  []string
)

func init() {
	pflag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	pflag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	pflag.StringSliceVar(&informers, "informer", []string{"deployments.v1.apps"}, "Resource to watch. This flag can be used multiple times.")
}

func main() {

	// Setup logging and parse flags
	klog.InitFlags(nil)
	pflag.Parse()

	// Create k8s client
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error creating dynamic client for config: %s", err.Error())
	}

	// Create & run a controller for each of the configured informers
	stopper := make(chan struct{})
	for _, informer := range informers {
		gvr, _ := schema.ParseResourceArg(informer)
		controller := controller.New(dc, gvr)
		go controller.Run(stopper)
	}

	// Block until we get signal to quick
	sigCh := make(chan os.Signal, 0)
	signal.Notify(sigCh, os.Kill, os.Interrupt)
	<-sigCh

	klog.Info("Server stopped")
}
