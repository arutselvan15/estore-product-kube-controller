package main

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	kubeclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	cc "github.com/arutselvan15/estore-common/clients"
	gc "github.com/arutselvan15/estore-common/config"
	"github.com/arutselvan15/estore-common/signals"
	"github.com/arutselvan15/estore-product-kube-client/pkg/client/clientset/versioned/scheme"
	pdtInformers "github.com/arutselvan15/estore-product-kube-client/pkg/client/informers/externalversions"

	cfg "github.com/arutselvan15/estore-product-kube-controller/config"
	"github.com/arutselvan15/estore-product-kube-controller/controllers"
	cLog "github.com/arutselvan15/estore-product-kube-controller/log"
)

func main() {
	var (
		config *rest.Config
		err    error
		log    = cLog.GetLogger()

		// set up signals so we handle the first shutdown signal gracefully
		stopCh = signals.SetupSignalHandler()
	)

	// kube config defined in env
	if gc.GetKubeConfigPath() != "" {
		config, err = clientcmd.BuildConfigFromFlags("", gc.GetKubeConfigPath())
		if err != nil {
			log.Errorf("error creating config using kube config path: %v", err.Error())
			os.Exit(cfg.ExitErrorCode)
		}
	} else {
		// default get current cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Errorf("error creating config using cluster config: %v", err)
			os.Exit(cfg.ExitErrorCode)
		}
	}

	// create client for config
	estoreClients, err := cc.NewEstoreClientForConfig(config)
	if err != nil {
		log.Errorf("error creating clients: %v", err)
		os.Exit(cfg.ExitErrorCode)
	}

	// product object related
	// retrieve our custom resource informer which was generated from the code generator and pass it the custom
	// resource client, specifying we should be looking through all namespaces for listing and watching
	pdtInformerFactory := pdtInformers.NewSharedInformerFactory(estoreClients.GetProductClient(), cfg.ResyncDuration)
	pdtInformer := pdtInformerFactory.Estore().V1().Products()

	// creating the rate limited work queue required for the controller
	pdtQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), cfg.WorkQueueName)

	// event broad caster
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(log.Infof)
	eventBroadcaster.StartRecordingToSink(&kubeclientv1.EventSinkImpl{
		Interface: estoreClients.GetKubeClient().CoreV1().Events(""),
	})

	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: fmt.Sprintf("%s-%s", cfg.ResourceName, cfg.Component)})

	pdtController := controllers.NewController(pdtInformer, pdtQueue, estoreClients, recorder, controllers.ProcessItem)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	pdtInformerFactory.Start(stopCh)

	pdtController.Run(cfg.WorkerCount, stopCh)
}
