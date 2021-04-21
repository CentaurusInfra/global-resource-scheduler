package main

import (
	"fmt"
	"flag"
	"k8s.io/klog"
    "time"
    "net/http"
	"io/ioutil"
	"sync"
	"encoding/json"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/apimachinery/pkg/types"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"gopkg.in/yaml.v2"
)

const (
	TimeOut = 10
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

type PodHandler struct {
	mu sync.Mutex // guards n
	n  int
	clientSet *kubernetes.Clientset
}

func NewPodHandler() *PodHandler {
	c, err := getClient()
	if err != nil {
		fmt.Printf("create pod handler error : %v", err)
		return nil
    }
	podHandler := &PodHandler {
		clientSet: c,
	}
	return podHandler
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

func (handler *PodHandler) getPod(w http.ResponseWriter, r *http.Request) (result string) {
	fmt.Println("server: get pod started")
	defer fmt.Println("server: get pod ended")
 	podName := ""
	for k, v := range r.URL.Query() {
		fmt.Printf("%s: %s\n", k, v)
		if(k == "name") {
			podName = v[0]
		}
    }
   
	//refer src/k8s.io/client-go/kubernetes/typed/core/v1
	if (podName == "") {
		options:= metav1.ListOptions{}
		pods, err := handler.clientSet.CoreV1().Pods(corev1.NamespaceDefault).List(options)
    	if err != nil {
			fmt.Println(err)
			result = Fail
		}
		strPods, err := yaml.Marshal(pods)
		if err != nil {
			result = Fail
		}
		result = string(strPods)
    } else {
		options:= metav1.GetOptions{}
		pod, err := handler.clientSet.CoreV1().Pods(corev1.NamespaceDefault).Get(podName,options)
		if err != nil {
			fmt.Println(err)
			result = Fail
		}
		strPod, err := yaml.Marshal(pod)
		if err != nil {
			result = Fail
		}
		result = string(strPod)
	}
	return result
}

func (handler *PodHandler) createPod(w http.ResponseWriter, r *http.Request) (result string) {
	fmt.Println("server: create pod started")
	defer fmt.Println("server: create pod ended")
	reqBody, err := ioutil.ReadAll(r.Body)
	fmt.Printf("createPod - reqBody: %v", reqBody)	
	if err != nil {
		klog.Errorf("request body error: %v", err)
	}
	pod := yaml2pod(reqBody)
	fmt.Printf("createPod - pod: %v", pod)	
    podCreated, err := handler.clientSet.CoreV1().Pods(corev1.NamespaceDefault).Create(&pod)
    if err != nil {
        fmt.Println(err)
        result = Fail
    }

	duration := int64(TimeOut * time.Second)
	status := podCreated.Status	
	options:= metav1.ListOptions{
		TimeoutSeconds: &duration,
		Watch: true,	
		ResourceVersion: podCreated.ResourceVersion,
		FieldSelector: 	 fmt.Sprintf("metadata.name=%s", pod.Name),
	}	
	watcher := handler.clientSet.CoreV1().Pods(corev1.NamespaceDefault).Watch(options)
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
			case <-time.After(TimeOut * time.Second):
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
}

func (handler *PodHandler) putPod(w http.ResponseWriter, r *http.Request) (result string) {
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

    podUpdated, err := handler.clientSet.CoreV1().Pods(corev1.NamespaceDefault).Update(&pod)
    if err != nil {
        fmt.Println(err)
        return
    }
	if (podUpdated.Status.Phase == corev1.ClusterScheduled) {
		fmt.Fprintf(w, "UpdatePod - Sucess- %v\n", podUpdated.Status.Phase)
	} else {
		select {
		case <-time.After(TimeOut * time.Second):
			options:= metav1.GetOptions{}
			podGet, _ := handler.clientSet.CoreV1().Pods(corev1.NamespaceDefault).Get(podName,options)
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
	strPod, err := yaml.Marshal(podUpdated)
	if err != nil {
		result = Fail
	}
	result = string(strPod)
	return result
}

func (handler *PodHandler) patchPod(w http.ResponseWriter, r *http.Request) (result string) {
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
	podPatched, err := handler.clientSet.CoreV1().Pods(corev1.NamespaceDefault).Patch(podName, types.ApplyPatchType, reqBody, "Spec", "Status")
	if (podPatched.Status.Phase == corev1.ClusterScheduled) {
		fmt.Fprintf(w, "UpdatePod - Sucess- %v\n", podPatched.Status.Phase)
	} else {
		select {
		case <-time.After(TimeOut * time.Second):
			options:= metav1.GetOptions{}
			podGet, _ := handler.clientSet.CoreV1().Pods(corev1.NamespaceDefault).Get(podPatched.Name,options)
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
	strPod, err := yaml.Marshal(podPatched)
	if err != nil {
		result = Fail
	}
	result = string(strPod)
	return result
}

func (handler *PodHandler) deletePod(w http.ResponseWriter, r *http.Request) (result string) {
	fmt.Println("server: delete pod started")
	defer fmt.Println("server: delete pod ended")
	var podName string
	for k, v := range r.URL.Query() {
		fmt.Printf("%s: %s\n", k, v)
		if(k == "name") {
			podName = v[0]
		}
    }
    // Get Pod by name
    options:= metav1.DeleteOptions{}
    err := handler.clientSet.CoreV1().Pods(corev1.NamespaceDefault).Delete(podName,&options)
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
	fmt.Printf("yaml2pod: %v", reqBody)
	if err := yaml.Unmarshal((reqBody), &body); err != nil {
        fmt.Println(err)
        return pod
	}

	//map to json structure
	jsonBody := convert(body)
	fmt.Printf("jsonBody: %v", jsonBody)
	//json structure to json string
	if b, err := json.Marshal(jsonBody); err != nil { 	
        klog.Errorf("request body marchal error: %v", err)
    } else {
		//json string to pod
		err = json.Unmarshal([]byte(b), &pod)
    }
	return pod
}

func (handler *PodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.n++
	fmt.Printf("http request: %v\n", r)
	if r.URL.Path != "/pods" {
		http.NotFound(w, r)
		return
	}
	var result string
	switch r.Method {
	case GET:
		result = handler.getPod(w, r)
    case POST:
		result = handler.createPod(w, r)
	case "PUT":
        result = handler.putPod(w, r)
	case "PATCH":
        result = handler.patchPod(w, r)
	case "DELETE":
        result = handler.deletePod(w, r)
    default:        
		result = http.StatusText(http.StatusNotImplemented)
		w.WriteHeader(http.StatusNotImplemented)
	} 
	w.Write([]byte(result))
}

//go run proxy.go url="localhost:8090"
//use example: 
//curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X POST http://localhost:8090/pods --data-binary @sample_1_pod.yaml
//curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X GET  http://localhost:8090/pods?name=pod-1
//curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X GET  http://localhost:8090/pods
//curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X DELETE  http://localhost:8090/pods?name=pod-1
//curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X PUT  http://localhost:8090/pods?name=pod-1 --data-binary @sample_1_pod.yaml
//curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X PATCH  http://localhost:8090/pods?name=pod-1 --data-binary @sample_1_pod.yaml
func main() {
	url := flag.String("url", ":8090", "proxy url")
	flag.Parse()
	podHandler := NewPodHandler()
	if(podHandler == nil) {
		klog.Error("cannot run http server - http handler is null")
	} else {
		http.Handle("/pods", podHandler)
		klog.Fatal(http.ListenAndServe(*url, nil))
	}
}
