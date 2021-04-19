package main

import (
	"fmt"
	"flag"
	"k8s.io/klog"
    "time"
    "net/http"
	"io/ioutil"
	"sync"
	// "reflect"
	"encoding/json"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/apimachinery/pkg/types"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	//appsv1 "k8s.io/api/apps/v1"
	// v1beta1 "k8s.io/api/extensions/v1beta1"
	// "k8s.io/apimachinery/pkg/api/errors"
	//"k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/util/diff"
	//"k8s.io/client-go/kubernetes/scheme"
	// "k8s.io/kubernetes/pkg/api"
    // _ "k8s.io/kubernetes/pkg/api/install"
    // _ "k8s.io/kubernetes/pkg/apis/extensions/install"
    // "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	//"k8s.io/apimachinery/pkg/labels"
	//typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"gopkg.in/yaml.v2"
)

const (
	TimeOut = 10 * time.Second
	Success	= "Success"
	Fail = "Fail"
	
	GET 	= "GET"
	POST	= "POST"
	LIST 	= "LIST"
	CREATE 	= "CREATE"
	PATCH 	= "PATCH"
	PUT 	= "PUT"
	DELETE 	= "DELETE"
)

type podHandler struct {
	mu sync.Mutex // guards n
	n  int
}

func getClient() (*kubernetes.Clientset, error) {
	kubeconfig := flag.String("kubeconfig", "/var/run/kubernetes/admin.kubeconfig", "Path to a kubeconfig. Only required if out-of-cluster.")
	masterURL := flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	config, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
    if err != nil {
      return nil, err
    }
    clientSet, err := kubernetes.NewForConfig(config)
    if err != nil {
		return nil, err
    }
	return clientSet, nil
}

//src/k8s.io/client-go/kubernetes/typed/core/v1
func getPod(w http.ResponseWriter, r *http.Request) (result string) {
	fmt.Println("server: get pod started")
	defer fmt.Println("server: get pod ended")
    ctx := r.Context()
	var podName string
	for k, v := range r.URL.Query() {
		fmt.Printf("%s: %s\n", k, v)
		if(k == "name") {
			podName = v[0]
		}
    }
	clientSet, err := getClient()
    // Get Pod by name
    options:= metav1.GetOptions{}
    pod, err := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Get(podName,options)
    if err != nil {
        fmt.Println(err)
		result = Fail
    }
	select {
	case <-time.After(TimeOut):
		options:= metav1.ListOptions{
			TimeoutSeconds: TimeOut,
			Watch: true			
		}	
		pod, err := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Get(podName,options)
		if (pod.Status.Phase == corev1.ClusterScheduled) {
			fmt.Fprintf(w, "GetPod - Sucess- %v\n", pod.Status.Phase)
			result = Success
		} else {
			fmt.Fprintf(w, "GetPod - Fail- %v\n", pod.Status.Phase)
			result = Fail
		}      	
	case <-ctx.Done():
		err := ctx.Err()
		fmt.Println("server:", err)
		internalError := http.StatusInternalServerError
		http.Error(w, err.Error(), internalError)
		result = Fail
	}
	return result
}

func createPod(w http.ResponseWriter, r *http.Request) (result string) {
	fmt.Println("server: create pod started")
	defer fmt.Println("server: create pod ended")
    ctx := r.Context()
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("request body error: %v", err)
	}
	pod := yaml2pod(reqBody)	
	clientSet, err := getClient()
    podCreated, err := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Create(&pod)
    if err != nil {
        fmt.Println(err)
        result = Fail
    }

	status := podCreated.Status	
	options:= metav1.ListOptions{
		TimeoutSeconds: TimeOut,
		Watch: true,	
		ResourceVersion: podCreated.ResourceVersion,
		FieldSelector:   fields.Set{"metadata.name": podName}.AsSelector(),
		LabelSelector:   labels.Everything(),		
	}	
	watcher, err := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Watch(podName,options)
	if (err != nil) {
		result = Fail
	}
	
	func() {
		for {
			select {
			case events, ok := <-watcher.ResultChan():
				if !ok {
					return
				}
				podCreated = events.Object.(*corev1.Pod)
				fmt.Println("Pod status:", podCreated.Status.Phase)
				status = podCreated.Status
				if podCreated.Status.Phase != corev1.ClusterScheduled {
					watcher.Stop()
				}
			case <-time.After(TimeOut):
				fmt.Println("timeout to wait for pod scheduled")
				watcher.Stop()
			}
		}
	}()
	if status.Phase != corev1.ClusterScheduled {
		fmt.Errorf("Pod is unavailable: %v", status.Phase)
		return Fail
	}
	return Success

	/*if (podCreated.Status.Phase == corev1.ClusterScheduled) {
		fmt.Fprintf(w, "GetPod - Sucess- %v\n", podCreated.Status.Phase)
	} else {
		select {
		case <-time.After(10 * time.Second):
			options:= metav1.ListOptions{}			
			podGet, _ := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Get(podCreated.Name,options)
			if (podGet.Status.Phase == corev1.ClusterScheduled) {
				fmt.Fprintf(w, "GetPod - Sucess- %v\n", podGet.Status.Phase)
			} else {
				fmt.Fprintf(w, "GetPod - Fail- %v\n", podGet.Status.Phase)
			}      	
		case <-ctx.Done():
			err := ctx.Err()
			fmt.Println("server:", err)
			internalError := http.StatusInternalServerError
			http.Error(w, err.Error(), internalError)
		}
	}*/
}

func putPod(w http.ResponseWriter, r *http.Request) (result string) {
	fmt.Println("server: put pod started")
	defer fmt.Println("server: put pod ended")
    ctx := r.Context()
	//request parameter
	var podName string
	for k, v := range r.URL.Query() {
		fmt.Printf("%s: %s\n", k, v)
		if(k == "name") {
			podName = v[0]
		}
    }
	//request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("request body error: %v", err)
	}
	pod := yaml2pod(reqBody)

	clientSet, err := getClient()
    podUpdated, err := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Update(&pod)
    if err != nil {
        fmt.Println(err)
        return
    }
	if (podUpdated.Status.Phase == corev1.ClusterScheduled) {
		fmt.Fprintf(w, "UpdatePod - Sucess- %v\n", podUpdated.Status.Phase)
	} else {
		select {
		case <-time.After(10 * time.Second):
			options:= metav1.GetOptions{}
			podGet, _ := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Get(podUpdated.Name,options)
			if (podGet.Status.Phase == corev1.ClusterScheduled) {
				fmt.Fprintf(w, "UpdatePod - Sucess- %v\n", podGet.Status.Phase)
			} else {
				fmt.Fprintf(w, "UpdatePod - Fail- %v\n", podGet.Status.Phase)
			}      	
		case <-ctx.Done():
			err := ctx.Err()
			fmt.Println("server:", err)
			internalError := http.StatusInternalServerError
			http.Error(w, err.Error(), internalError)
		}
	}
}

func patchPod(w http.ResponseWriter, r *http.Request) (result string) {
	fmt.Println("server: put pod started")
	defer fmt.Println("server: put pod ended")
    ctx := r.Context()
	//request parameter
	var podName string
	for k, v := range r.URL.Query() {
		fmt.Printf("%s: %s\n", k, v)
		if(k == "name") {
			podName = v[0]
		}
    }
	//request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("request body error: %v", err)
	}
	pod := yaml2pod(reqBody)

	clientSet, err := getClient()
	options:= metav1.GetOptions{}
    podOld, err := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Get(podName, options)
    if err != nil {
        fmt.Println(err)
        return
    }
	podPatched, err := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Patch(podName, types.ApplyPatchType, reqBody, "Spec", "Status")
    // Print its creation time
    fmt.Println(podPatched.GetCreationTimestamp())
	if (podPatched.Status.Phase == corev1.ClusterScheduled) {
		fmt.Fprintf(w, "UpdatePod - Sucess- %v\n", podPatched.Status.Phase)
	} else {
		select {
		case <-time.After(10 * time.Second):
			options:= metav1.GetOptions{}
			podGet, _ := clientSet.CoreV1().Pods(corev1.NamespaceDefault).Get(podPatched.Name,options)
			if (podGet.Status.Phase == corev1.ClusterScheduled) {
				fmt.Fprintf(w, "UpdatePod - Sucess- %v\n", podGet.Status.Phase)
			} else {
				fmt.Fprintf(w, "UpdatePod - Fail- %v\n", podGet.Status.Phase)
			}      	
		case <-ctx.Done():
			err := ctx.Err()
			fmt.Println("server:", err)
			internalError := http.StatusInternalServerError
			http.Error(w, err.Error(), internalError)
		}
	}
}

func deletePod(w http.ResponseWriter, r *http.Request) (result string) {
	fmt.Println("server: delete pod started")
	defer fmt.Println("server: delete pod ended")
    ctx := r.Context()
	var podName string
	for k, v := range r.URL.Query() {
		fmt.Printf("%s: %s\n", k, v)
		if(k == "name") {
			podName = v[0]
		}
    }
	clientSet, err := getClient()
    // Get Pod by name
    options:= metav1.DeleteOptions{}
    err = clientSet.CoreV1().Pods(corev1.NamespaceDefault).Delete(podName,&options)
    if err != nil {
        fmt.Println(err)
        result = Fail  
    }
	result = Success
	return 
}

//convert map to json 
func convert(i interface{}) interface{} {
    switch x := i.(type) {
    case map[interface{}]interface{}:
        m2 := map[string]interface{}{}
        for k, v := range x {
            m2[k.(string)] = convert(v)
        }
        return m2
    case []interface{}:
        for i, v := range x {
            x[i] = convert(v)
        }
    }
    return i
}

func yaml2pod(reqBody []byte) (corev1.Pod) {
	var pod corev1.Pod
	var body interface{}
	if err := yaml.Unmarshal([]byte(reqBody), &body); err != nil {
        fmt.Println(err)
        return pod
	}
	//map to json structure
	jsonBody := convert(body)
	//json structure to json string
	if b, err := json.Marshal(jsonBody); err != nil { 	
        klog.Errorf("request body marchal error: %v", err)
    } else {
		//json string to pod
		err = json.Unmarshal([]byte(b), &pod)
    }
	return pod
}

func (handler *podHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.n++
	fmt.Printf("http request: %v\n", r)
	if r.URL.Path != "/pods" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case GET:
		result = getPod(w, r)
    case POST:
		result = createPod(w, r)
	case "PUT":
        result = putPod(w, r)
	case "PATCH":
        result = patchPod(w, r)
	case "DELETE":
        result = deletePod(w, r)
    default:        
		result = http.StatusText(http.StatusNotImplemented)
		w.WriteHeader(http.StatusNotImplemented)
	} 
	w.Write([]byte(result))
}

//go run proxy.go url="localhost:8090"
func main() {
	url := flag.String("url", ":8090", "proxy url")
	flag.Parse()
	http.Handle("/pods", new(podHandler))
	klog.Fatal(http.ListenAndServe(url, nil))
}
