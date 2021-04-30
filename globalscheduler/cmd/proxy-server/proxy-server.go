/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	TimeOut = 10
	Success = "Success"
	Fail    = "Fail"

	GET    = "GET"
	POST   = "POST"
	LIST   = "LIST"
	CREATE = "CREATE"
	PATCH  = "PATCH"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type PodHandler struct {
	mu        sync.Mutex
	namespace string
	clientSet *kubernetes.Clientset
}

func NewPodHandler(ns string) *PodHandler {
	c, err := getClient()
	if err != nil {
		klog.Errorf("create a new pod handler error : %v", err)
		return nil
	}
	podHandler := &PodHandler{
		namespace: ns,
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
	klog.Info("server: get pod started")
	defer klog.Info("server: get pod ended")
	podName := ""
	for k, v := range r.URL.Query() {
		if k == "name" {
			podName = v[0]
		}
	}
	if podName == "" {
		options := metav1.ListOptions{}
		pods, err := handler.clientSet.CoreV1().Pods(handler.namespace).List(options)
		if err != nil {
			klog.Errorf("pod list error: %v", err)
			result = Fail
		}
		strPods, err := yaml.Marshal(pods)
		if err != nil {
			result = Fail
		}
		result = string(strPods)
	} else {
		options := metav1.GetOptions{}
		pod, err := handler.clientSet.CoreV1().Pods(handler.namespace).Get(podName, options)
		if err != nil {
			klog.Errorf("pod get error: %v", err)
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
	klog.Info("server: create pod started")
	defer klog.Info("server: create pod ended")
	ctx := r.Context()
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("request body error: %v", err)
	}
	pod, err := yaml2pod(reqBody)
	if err != nil {
		klog.Errorf("error while converting yaml to pod: %v", err)
		result = Fail
		return result
	}
	podCreated, err := handler.clientSet.CoreV1().Pods(handler.namespace).Create(&pod)
	if err != nil {
		klog.Errorf("create pod error: %v", err)
		result = Fail
		return result
	}
	duration := int64(TimeOut * time.Second * 2)
	status := string(podCreated.Status.Phase)
	options := metav1.ListOptions{
		TimeoutSeconds:  &duration,
		Watch:           true,
		ResourceVersion: podCreated.ResourceVersion,
		FieldSelector:   fmt.Sprintf("metadata.name=%s,status.phase=%s", pod.Name, corev1.ClusterScheduled), //PodBound, ClusterScheduled
	}
	watcher := handler.clientSet.CoreV1().Pods(handler.namespace).Watch(options)
	timer := time.NewTimer(TimeOut * time.Second)
	func() {
		defer watcher.Stop()
		bLoop := true
		for bLoop {
			select {
			case event := <-watcher.ResultChan():
				podObject, ok := event.Object.(*corev1.Pod)
				if ok {
					status = string(podObject.Status.Phase)
					if status == string(corev1.ClusterScheduled) {
						bLoop = false
					}
				}
			case <-timer.C:
				klog.Info("timeout to wait for scheduling a pod & delete the pod ")
				options := metav1.DeleteOptions{}
				err := handler.clientSet.CoreV1().Pods(handler.namespace).Delete(pod.Name, &options)
				if err != nil {
					klog.Errorf("pod deletion error after failing pod creation: %v", err)
					result = Fail
				}
				bLoop = false
			case <-ctx.Done():
				err := ctx.Err()
				if err != nil {
					klog.Errorf("server error: %v", err)
				}
				internalError := http.StatusInternalServerError
				http.Error(w, err.Error(), internalError)
				bLoop = false
			}
		}
	}()

	if status != string(corev1.ClusterScheduled) {
		klog.Errorf("Pod scheduling is failed: %v", status)
		return Fail
	}
	klog.Infof("Pod scheduling is succeeded: %v", status)
	return Success
}

func (handler *PodHandler) putPod(w http.ResponseWriter, r *http.Request) (result string) {
	klog.Info("server: put pod started")
	defer klog.Info("server: put pod ended")
	ctx := r.Context()
	//request parameter
	var podName string
	for k, v := range r.URL.Query() {
		if k == "name" {
			podName = v[0]
		}
	}
	//request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("request body error: %v", err)
	}
	pod, err := yaml2pod(reqBody)
	if err != nil {
		klog.Errorf("error while converting yaml to pod: %v", err)
		result = Fail
		return result
	}

	podUpdated, err := handler.clientSet.CoreV1().Pods(handler.namespace).Update(&pod)
	if err != nil {
		klog.Errorf("update pod error: %v", err)
		result = Fail
		return result
	}

	duration := int64(TimeOut * time.Second * 2)
	status := string(podUpdated.Status.Phase)
	options := metav1.ListOptions{
		TimeoutSeconds:  &duration,
		Watch:           true,
		ResourceVersion: podUpdated.ResourceVersion,
		FieldSelector:   fmt.Sprintf("metadata.name=%s,status.phase=%s", podName, corev1.ClusterScheduled), //PodBound, ClusterScheduled
	}
	watcher := handler.clientSet.CoreV1().Pods(handler.namespace).Watch(options)
	timer := time.NewTimer(TimeOut * time.Second)
	func() {
		defer watcher.Stop()
		bLoop := true
		for bLoop {
			select {
			case event := <-watcher.ResultChan():
				podObject, ok := event.Object.(*corev1.Pod)
				if ok {
					status = string(podObject.Status.Phase)
					if status == string(corev1.ClusterScheduled) {
						bLoop = false
					}
				}
			case <-timer.C:
				klog.Info("timeout to wait for updating a pod")
				bLoop = false
			case <-ctx.Done():
				err := ctx.Err()
				if err != nil {
					klog.Errorf("server error: %v", err)
				}
				internalError := http.StatusInternalServerError
				http.Error(w, err.Error(), internalError)
			}
		}
	}()

	if status != string(corev1.ClusterScheduled) {
		klog.Errorf("Pod update is failed: %v", status)
		return Fail
	}
	klog.Infof("Pod update is succeeded: %v", status)
	return Success
}

func (handler *PodHandler) patchPod(w http.ResponseWriter, r *http.Request) (result string) {
	klog.Info("server: patch pod started")
	defer klog.Info("server: patch pod ended")
	ctx := r.Context()
	//request parameter
	var podName string
	for k, v := range r.URL.Query() {
		if k == "name" {
			podName = v[0]
		}
	}
	//request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("request body error: %v", err)
	}
	jsonstr, err := yaml2json(reqBody)
	if err != nil {
		klog.Errorf("error while converting yaml to pod: %v", err)
		result = Fail
		return
	}
	podPatched, err := handler.clientSet.CoreV1().Pods(handler.namespace).Patch(podName, types.MergePatchType, []byte(jsonstr))
	if err != nil {
		klog.Errorf("patch pod error: %v", err)
		result = Fail
		return result
	}

	duration := int64(TimeOut * time.Second)
	status := string(podPatched.Status.Phase)
	options := metav1.ListOptions{
		TimeoutSeconds:  &duration,
		Watch:           true,
		ResourceVersion: podPatched.ResourceVersion,
		FieldSelector:   fmt.Sprintf("metadata.name=%s,status.phase=%s", podName, corev1.ClusterScheduled), //PodBound, ClusterScheduled
	}
	watcher := handler.clientSet.CoreV1().Pods(handler.namespace).Watch(options)
	timer := time.NewTimer(TimeOut * time.Second)
	func() {
		defer watcher.Stop()
		bLoop := true
		for bLoop {
			select {
			case event := <-watcher.ResultChan():
				podObject, ok := event.Object.(*corev1.Pod)
				if ok {
					status = string(podObject.Status.Phase)
					if status == string(corev1.ClusterScheduled) {
						bLoop = false
					}
				}
			case <-timer.C:
				klog.Info("timeout to wait for patching a pod")
				bLoop = false
			case <-ctx.Done():
				err := ctx.Err()
				if err != nil {
					klog.Errorf("server error: %v", err)
				}
				internalError := http.StatusInternalServerError
				http.Error(w, err.Error(), internalError)
				bLoop = false
			}
		}
	}()
	if status != string(corev1.ClusterScheduled) {
		klog.Errorf("Pod patch is failed: %v", status)
		return Fail
	}
	klog.Infof("Pod patch is succeeded: %v", status)
	return Success
}

func (handler *PodHandler) deletePod(w http.ResponseWriter, r *http.Request) (result string) {
	klog.Info("server: delete pod started")
	defer klog.Info("server: delete pod ended")
	var podName string
	for k, v := range r.URL.Query() {
		if k == "name" {
			podName = v[0]
		}
	}
	// Get Pod by name
	options := metav1.DeleteOptions{}
	err := handler.clientSet.CoreV1().Pods(handler.namespace).Delete(podName, &options)
	if err != nil {
		klog.Errorf("delete pod error: %v", err)
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

func yaml2pod(reqBody []byte) (pod corev1.Pod, err error) {
	//yaml to json string
	if str, err := yaml2json(reqBody); err != nil {
		klog.Errorf("request body marchal error: %v", err)
	} else {
		//json string to pod
		err = json.Unmarshal([]byte(str), &pod)
	}
	return pod, err
}

func yaml2json(reqBody []byte) (str string, err error) {
	var body interface{}
	str = ""
	if err := yaml.Unmarshal((reqBody), &body); err != nil {
		klog.Errorf("yaml unmarshalling error: %v", err)
		return str, err
	}
	//map to json structure
	jsonBody := convert(body)
	//json structure to json string
	b, err := json.Marshal(jsonBody)
	if err != nil {
		klog.Errorf("request body marchalling error: %v", err)
	}
	str = string(b)
	return str, err
}

func (handler *PodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
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

// go run proxy-server.go url="localhost:8090"
//use example:
//curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X POST http://localhost:8090/pods --data-binary @sample_1_pod.yaml
// curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X GET  http://localhost:8080/globalpods?name=pod-1
// curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X GET  http://localhost:8090/pods
// curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X PUT  http://localhost:8090/pods?name=pod-1 --data-binary @sample_1_pod.yaml
// curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X PATCH  http://localhost:8090/pods?name=pod-1 --data-binary @sample_1_pod.yaml
// curl -H "Accept: application/x-yaml" -H "Content-Type: application/x-yaml" -X DELETE  http://localhost:8090/pods?name=pod-1
func main() {
	url := flag.String("url", ":8090", "proxy url")
	namespace := flag.String("namespace", "default", "namespace")
	flag.Parse()
	podHandler := NewPodHandler(*namespace)
	if podHandler == nil {
		klog.Error("cannot run http server - http handler is null")
	} else {
		http.Handle("/pods", podHandler)
		klog.Fatal(http.ListenAndServe(*url, nil))
	}
}
