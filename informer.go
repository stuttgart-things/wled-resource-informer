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
	"sync"
	"time"

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
	nodeName    = os.Getenv("NODE_NAME")
	wledSegment = os.Getenv(nodeName)
	wledUrl     = os.Getenv("WLED_URL")
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

	// listOp := metav1.ListOptions{
	// 	FieldSelector: "spec.nodeName=pve-dev-2",
	// }

	// listOpfunc := dynamicinformer.TweakListOptionsFunc(func(options *metav1.ListOptions) { *options = listOp })
	// factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, corev1.NamespaceAll, listOpfunc)

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, corev1.NamespaceAll, nil)

	informer := factory.ForResource(resource).Informer()

	mux := &sync.RWMutex{}
	synced := false

	// START INFORMING

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			fmt.Println("ADDED POD ON NODE: " + nodeName)
			fmt.Println("WOULD SEND TO", wledSegment)

			wled.ControllWled(wledUrl)

			// fmt.Println(obj)

			// CONVERT OBJECT TO POD
			createdUnstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
			fmt.Println(err)

			po := new(corev1.Pod)

			err = runtime.DefaultUnstructuredConverter.FromUnstructured(createdUnstructuredObj, &po)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("POD", po.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			fmt.Println("UPDATED!")
			fmt.Println("UPDATED POD ON NODE: " + os.Getenv("NODE_NAME"))

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
