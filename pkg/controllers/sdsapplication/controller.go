package sdsapplication

import (
	"k8s.io/apimachinery/pkg/fields"

	api "github.com/samsung-cnct/cma-operator/pkg/apis/cma/v1alpha1"
	"github.com/samsung-cnct/cma-operator/pkg/generated/cma/client/clientset/versioned"
	"github.com/samsung-cnct/cma-operator/pkg/util/k8sutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

type SDSApplicationController struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller

	client *versioned.Clientset
}

func NewSDSApplicationController(config *rest.Config) (output *SDSApplicationController) {
	if config == nil {
		config = k8sutil.DefaultConfig
	}
	client := versioned.NewForConfigOrDie(config)

	// create sdsapplication list/watcher
	sdsApplicationListWatcher := cache.NewListWatchFromClient(
		client.CmaV1alpha1().RESTClient(),
		api.SDSApplicationResourcePlural,
		"default",
		fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the SDSCluster key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the SDSPackageManager than the version which was responsible for triggering the update.
	//indexer, informer := cache.NewIndexerInformer(
	//	sdsApplicationListWatcher,
	//	&api.SDSApplication{},
	//	30*time.Second,
	//	cache.ResourceEventHandlerFuncs{
	//		AddFunc: func(obj interface{}) {
	//			key, err := cache.MetaNamespaceKeyFunc(obj)
	//			if err == nil {
	//				queue.Add(key)
	//			}
	//		},
	//		UpdateFunc: func(old interface{}, new interface{}) {
	//			key, err := cache.MetaNamespaceKeyFunc(new)
	//			if err == nil {
	//				queue.Add(key)
	//			}
	//		},
	//		DeleteFunc: func(obj interface{}) {
	//			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
	//			// key function.
	//			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	//			if err == nil {
	//				queue.Add(key)
	//			}
	//		},
	//	}, cache.Indexers{})

	sharedInformer := cache.NewSharedIndexInformer(
		sdsApplicationListWatcher,
		&api.SDSApplication{},
		30*time.Second,
		cache.Indexers{},
	)

	sharedInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			logger.Infof("Add")
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			logger.Infof("Update")
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			logger.Infof("Delete")
			if err == nil {
				queue.Add(key)
			}
		},
	}, 30*time.Second)

	output = &SDSApplicationController{
		informer: sharedInformer,
		indexer:  sharedInformer.GetIndexer(),
		//informer: informer,
		//indexer:  indexer,
		queue:  queue,
		client: client,
	}
	output.SetLogger()
	return
}
