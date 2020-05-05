package main

import (
	"flag"
	"fmt"
	"github.com/amimof/synka/pkg/controller"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"os"
	"os/signal"
	"syscall"
)

var (
	// VERSION of the app. Is set when project is built and should never be set manually
	VERSION string
	// COMMIT is the Git commit currently used when compiling. Is set when project is built and should never be set manually
	COMMIT string
	// BRANCH is the Git branch currently used when compiling. Is set when project is built and should never be set manually
	BRANCH string
	// GOVERSION used to compile. Is set when project is built and should never be set manually
	GOVERSION string

	config               string
	masterURL            string
	kubeconfig           string
	informers            []string
	onlyOneSignalHandler = make(chan struct{})
	shutdownSignals      = []os.Signal{os.Kill, os.Interrupt, syscall.SIGTERM}
	defaultInformers     = []string{"deployments.v1.apps", "pods.v1.", "namespaces.v1.", "services.v1.", "serviceaccounts.v1."}
)

func init() {
	pflag.StringVar(&config, "config", "/etc/synka/config.yaml", "Path to synka configuration file.")
	pflag.StringVar(&kubeconfig, "kubeconfig", "~/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster. Synka synchronises configuration from the current-context defined in this file to contexts in kubeconfig defined by --config.")
	pflag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	pflag.StringSliceVar(&informers, "informer", defaultInformers, "Resource to watch. This flag can be used multiple times.")
}

// setupSignalHandler returns a stop channel which is closed when program receives a SIGKILL, SIGINT or SIGTERM.
// If a second signal is caught, the program is terminated with exit code 1.
func setupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

func setupConfig() (*controller.Config, error) {
	viper.SetConfigFile(config)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var config *controller.Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func main() {

	// Setup version flag
	showver := pflag.Bool("version", false, "Print version")

	// Setup our usage func
	pflag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage:\n")
		fmt.Fprint(os.Stderr, "  synka [OPTIONS]\n\n")
		fmt.Fprint(os.Stderr, "Synka synchronizes Kubernetes state between clusters https://synka.dev\n\n")
		fmt.Fprintln(os.Stderr, pflag.CommandLine.FlagUsages())
	}

	// Setup logging and parse flags
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	c, err := setupConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// Show version if requested
	if *showver {
		fmt.Printf("Version: %s\nCommit: %s\nBranch: %s\nGoVersion: %s\n", VERSION, COMMIT, BRANCH, GOVERSION)
		return
	}

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
	stopCh := setupSignalHandler()
	for _, informer := range informers {
		gvr, _ := schema.ParseResourceArg(informer)
		controller := controller.New(dc, c, gvr)
		go controller.Run(stopCh)
	}

	// Block until we get signal to quit
	<-stopCh
	klog.Info("Server stopped")

}
