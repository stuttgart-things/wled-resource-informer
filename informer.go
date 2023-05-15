/*
Copyright © 2023 Patrick Hermann patrick.hermann@sva.de
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	wled "github.com/stuttgart-things/wled-resource-informer/wled"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	nodeName          = os.Getenv("NODE_NAME")
	wledSegment       = os.Getenv(nodeName)
	brightness        = os.Getenv("BRIGHTNESS")
	wledUrl           = os.Getenv("WLEDURL")
	informerNamespace = os.Getenv("NAMESPACE")
)

func main() {

	// KUBECONFIG HANDLING OUTSIDE/INSIDE CLUSTER
	kubeConfig := os.Getenv("KUBECONFIG")

	var clusterConfig *rest.Config
	var err error
	if kubeConfig != "" {
		clusterConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	} else {
		clusterConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Fatalln(err)
	}

	clusterClient, err := dynamic.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalln(err)
	}

	// INFORMER CONFIG

	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	listOp := metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	}

	listOpfunc := dynamicinformer.TweakListOptionsFunc(func(options *metav1.ListOptions) { *options = listOp })
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, corev1.NamespaceAll, listOpfunc)

	// factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, corev1.NamespaceAll, nil)

	if informerNamespace != "" {
		factory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, informerNamespace, listOpfunc)
	}

	informer := factory.ForResource(resource).Informer()

	mux := &sync.RWMutex{}
	synced := false

	// START INFORMING

	fmt.Println("INFORMING ON: " + nodeName)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			fmt.Println("ADDED POD ON NODE: " + nodeName)
			fmt.Println("LED SEGMENT", wledSegment)
			fmt.Println("NAMESPACE", informerNamespace)

			// CONVERT OBJECT TO POD
			createdUnstructuredObj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)

			po := new(corev1.Pod)

			err = runtime.DefaultUnstructuredConverter.FromUnstructured(createdUnstructuredObj, &po)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("POD", po.Name)

			updatedStatus := wled.WledStatus{
				Brightness: convertStringToInt(brightness),
				Segment:    convertStringToInt(wledSegment),
				Color:      "[145,92,210],[12,32,0],[0,0,0]",
				Fx:         0,
			}

			wled.ControllWled(wledUrl, updatedStatus)

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			// fmt.Println("UPDATED!")
			// fmt.Println("UPDATED POD ON NODE: " + os.Getenv("NODE_NAME"))

			// ControllWled()
			// Handler logic
		},
		DeleteFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			fmt.Println("DELETED!")
			fmt.Println("DELETED POD ON NODE: " + os.Getenv("NODE_NAME"))

		},
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go informer.Run(ctx.Done())

	isSynced := cache.WaitForCacheSync(ctx.Done(), informer.HasSynced)
	mux.Lock()
	synced = isSynced
	mux.Unlock()

	if !isSynced {
		log.Fatal("failed to sync")
	}

	<-ctx.Done()
}

func convertStringToInt(input string) (output int) {
	output, err := strconv.Atoi(wledSegment)
	if err != nil {
		log.Fatal(err)
	}

	return

}
