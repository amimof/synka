package controller

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"time"
)

// Controller is a k8s controller implementation
type Controller struct {
	queue    workqueue.RateLimitingInterface
	factory  dynamicinformer.DynamicSharedInformerFactory
	gvr      *schema.GroupVersionResource
	indexer  cache.Indexer
	clusters []Cluster
	config   *Config
}

// New creates a new instance of controller for the given GroupVersionResource
// See https://godoc.org/k8s.io/apimachinery/pkg/runtime/schema#GroupVersionResource for more information
func New(client dynamic.Interface, config *Config, gvr *schema.GroupVersionResource) *Controller {
	return &Controller{
		factory: dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, v1.NamespaceAll, nil),
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "synka.io"),
		gvr:     gvr,
		config:  config,
	}
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh channel
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()
	informer := c.factory.ForResource(*c.gvr)
	c.indexer = informer.Informer().GetIndexer()
	go c.startWatching(stopCh, informer.Informer())

	if !cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	go wait.Until(c.runWorker, time.Second, stopCh)

	klog.Infof("Started controller for %s", c.gvr.GroupResource().String())
	<-stopCh
	klog.Infof("Shutting down controller for %s", c.gvr.GroupResource().String())

}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.syncToStdout(key.(string))
	c.handleErr(err, key)
	return true
}

func (c *Controller) syncToStdout(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)

	if err != nil {
		klog.Errorf("Fetching objects witch key %s from store failed with %v", key, err)
		return err
	}

	u := obj.(*unstructured.Unstructured)

	if !exists {
		klog.Infof("Resource %s does not exists anymore", key)
	} else {

		for _, cluster := range c.config.Clusters {

			// Get a client for the GroupVersionResource
			client, err := cluster.GetClient(c.gvr)
			if err != nil {
				return err
			}

			// Only go any further if object is annotated properly
			a := u.GetAnnotations()
			if _, ok := a["synka.io/sync"]; ok {
				if a["synka.io/sync"] == "true" {

					klog.V(4).Infof("Sync %s/%s: %s", u.GetAPIVersion(), u.GetKind(), u.GetName())

					// Remove any immutable fields
					u.Object["metadata"].(map[string]interface{})["resourceVersion"] = ""

					// Create the resource in target cluster using the cluters's dynamic client
					result, err := client.Resource(*c.gvr).Namespace(u.GetNamespace()).Create(context.Background(), u, v1.CreateOptions{})
					if err != nil {
						return err
					}
					klog.V(4).Infof("Created %s/%s on %s", result.GetKind(), result.GetName(), cluster.Name)

				}
			}
		}
	}

	return nil
}

func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}
	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing resource %s: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}
	c.queue.Forget(key)
	runtime.HandleError(err)
	klog.Infof("Dropping resource %s out of the queue: %v", key, err)
}

func (c *Controller) startWatching(stopCh <-chan struct{}, s cache.SharedIndexInformer) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// u := obj.(*unstructured.Unstructured)
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			//u := obj.(*unstructured.Unstructured)
			//klog.Infof("received update event! %s", u.GroupVersionKind())
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			//u := obj.(*unstructured.Unstructured)
			//klog.Infof("received update event! %s", u.GroupVersionKind())
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
	}

	s.AddEventHandler(handlers)
	s.Run(stopCh)
}
