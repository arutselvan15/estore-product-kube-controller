// Package controllers controllers
package controllers

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	cc "github.com/arutselvan15/estore-common/clients"
	pdtv1 "github.com/arutselvan15/estore-product-kube-client/pkg/apis/estore/v1"
	pdtv1Informers "github.com/arutselvan15/estore-product-kube-client/pkg/client/informers/externalversions/estore/v1"

	cfg "github.com/arutselvan15/estore-product-kube-controller/config"
	gLog "github.com/arutselvan15/estore-product-kube-controller/log"
)

var log = gLog.GetLogger()

// Controller controller
type Controller struct {
	pdtInformer     cache.SharedIndexInformer
	pdtListerSynced cache.InformerSynced
	pdtQueue        workqueue.RateLimitingInterface
	clients         cc.EstoreClientInterface
	recorder        record.EventRecorder
	processItem     ProcessItemType
}

// NewController new controller
func NewController(pdtInformer pdtv1Informers.ProductInformer, pdtQueue workqueue.RateLimitingInterface,
	clients cc.EstoreClientInterface, recorder record.EventRecorder, processItem ProcessItemType) *Controller {
	c := &Controller{
		pdtInformer:     pdtInformer.Informer(),
		pdtListerSynced: pdtInformer.Informer().HasSynced,
		pdtQueue:        pdtQueue,
		clients:         clients,
		recorder:        recorder,
		processItem:     processItem,
	}

	pdtInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if key := onAdd(obj.(*pdtv1.Product), c.recorder); key != "" {
					c.pdtQueue.Add(key)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if key := onUpdate(oldObj.(*pdtv1.Product), newObj.(*pdtv1.Product), c.recorder); key != "" {
					c.pdtQueue.Add(key)
				}
			},
			DeleteFunc: func(obj interface{}) {
				if key := onDelete(obj.(*pdtv1.Product), c.recorder); key != "" {
					c.pdtQueue.Add(key)
				}
			},
		},
	)

	return c
}

// Run run the controller with number of workers mentioned in the arguments
func (c *Controller) Run(workerCount int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()

	// let the workers stop when we are done
	defer c.pdtQueue.ShutDown()

	// waits for caches to populate.  It returns true if it was successful, false if the controller should shutdown
	// wait for the caches to synchronize before starting the workers
	if !cache.WaitForNamedCacheSync(fmt.Sprintf("%s-%s", cfg.ResourceName, cfg.Component), stopCh, c.pdtListerSynced) {
		log.Error("timed out waiting for caches to sync product resource")
		return
	}

	// launch worker(s) to process the resources
	for i := 0; i < workerCount; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup something in a cache
	key, shutdown := c.pdtQueue.Get()
	if shutdown {
		return false
	}

	// tell the queue that we are done with processing this key. This unblocks the key for other workers
	// this allows safe parallel processing because two pods with the same key are never processed in parallel.
	defer c.pdtQueue.Done(key)

	err := c.doSync(key.(string))
	if err != nil {
		// re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		// you can custom logic here to take decision to re-process the item or not
		c.pdtQueue.AddRateLimited(key)
		runtime.HandleError(fmt.Errorf("doSync failed for key %s, error: %v", key, err))
	} else {
		// forget about the #AddRateLimited history of the key on every successful synchronization.
		// this ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.pdtQueue.Forget(key)
	}

	return true
}

func (c *Controller) doSync(key string) error {
	pdt := &pdtv1.Product{}

	obj, exists, err := c.pdtInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("fetching object with key %s from cache failed with %v", key, err)

		return err
	}

	if !exists {
		return nil
	}

	if obj != nil {
		pdt = obj.(*pdtv1.Product)
	}

	return c.processItem(pdt, c.clients, c.recorder)
}
