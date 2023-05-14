package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
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
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			fmt.Println("ADDED!")
			fmt.Println(obj)

			createdUnstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
			fmt.Println(err)

			po := new(corev1.Pod)

			err = runtime.DefaultUnstructuredConverter.FromUnstructured(createdUnstructuredObj, &po)
			if err != nil {
				log.Fatal(err)
			}

			// fmt.Println("NODE", po.Node)
			// Handler logic
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			fmt.Println("UPDATED!")
			fmt.Println(RandBool)
			// ControllLights()
			// Handler logic
		},
		DeleteFunc: func(obj interface{}) {
			mux.RLock()
			defer mux.RUnlock()
			if !synced {
				return
			}

			fmt.Println("DELETED!")
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

func ControllLights() {

	httpposturl := "http://wled-87552c.local/json/state"
	fmt.Println("HTTP JSON POST URL:", httpposturl)

	var jsonData = []byte(`{
	"on": "t",
	"v": true
		}`)

	request, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))
}

func RandBool() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(2) == 1
}
