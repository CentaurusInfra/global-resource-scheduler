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

package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"net/http"
)

type server struct {
	Name           string              `json:"name"`
	ImageRef       string              `json:"imageRef"`
	FlavorRef      string              `json:"flavorRef"`
	Networks       []map[string]string `json:"networks"`
	SecurityGroups []map[string]string `json:"security_groups"`

	// Display the name of a cluster
	Metadata map[string]string `json:"metadata"`
}

/*
	DeleteInstance: delete openstack server with instanceID
	Input:
		host: Private IPv4 addresses
		authToken: the given token string
		instanceID: instance identification string
	Output:
		return nil if delete success, otherwise delete fail with error message
*/

func DeleteInstance(host string, authToken string, instanceID string) error {
	instanceDetailsURL := "http://" + host + "/compute/v2.1/servers/" + instanceID
	req, _ := http.NewRequest("DELETE", instanceDetailsURL, nil)
	req.Header.Set("X-Auth-Token", authToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.V(3).Infof("HTTP DELETE Instance Status Request Failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	// http.StatusNoContent = 204
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("bad request for HTTP DELETE instance request")
	} else {
		klog.V(3).Infof("HTTP DELETE Instance Status Request Success")
		return nil
	}
}

/*
	ServerCreate: create openstack server
	Input:
		host: Private IPv4 addresses
		authToken: the given token string
		manifest: pod spec structure
	Output:
		return instanceID if no error, otherwise return error
*/
func ServerCreate(host string, authToken string, manifest *v1.PodSpec) (string, error) {
	serverCreateRequestURL := "http://" + host + "/compute/v2.1/servers"
	serverStruct := server{
		Name:      manifest.VirtualMachine.Name,
		ImageRef:  manifest.VirtualMachine.Image,
		FlavorRef: manifest.VirtualMachine.Flavors[0].FlavorID,
		Networks: []map[string]string{
			{"uuid": manifest.Nics[0].Name},
		},
		SecurityGroups: []map[string]string{
			{"name": manifest.VirtualMachine.SecurityGroupId},
		},
		Metadata: map[string]string{
			"ClusterName": manifest.ClusterName,
		},
	}
	serverJson := map[string]server{}
	serverJson["server"] = serverStruct
	finalData, _ := json.Marshal(serverJson)
	req, _ := http.NewRequest("POST", serverCreateRequestURL, bytes.NewBuffer(finalData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", authToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.V(3).Infof("HTTP Post Instance Request Failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var instanceResponse map[string]interface{}
	if err := json.Unmarshal(body, &instanceResponse); err != nil {
		klog.V(3).Infof("Instance Create Response Unmarshal Failed")
		return "", err
	}

	// http.StatusForbidden = 403
	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("instance capacity has reached its limit")
	}

	if instanceResponse["server"] == nil {
		return "", fmt.Errorf("bad request for server create")
	}
	serverResponse := instanceResponse["server"].(map[string]interface{})
	instanceID := serverResponse["id"].(string)

	return instanceID, nil
}

/*
	TokenExpired: Identify if the given token has been expired
	Input:
		host: Private IPv4 addresses
		authToken: the given token string
	Output:
		if return true, then token has been expired
		if return false, then token can still be used
*/
func TokenExpired(host string, authToken string) bool {
	checkTokenURL := "http://" + host + "/identity/v3/auth/tokens"
	req, _ := http.NewRequest("HEAD", checkTokenURL, nil)
	req.Header.Set("X-Auth-Token", authToken)
	req.Header.Set("X-Subject-Token", authToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.V(3).Infof("HTTP Check Token Request Failed: %v", err)
	}
	defer resp.Body.Close()

	// http.StatusOK = 200
	if resp.StatusCode != http.StatusOK {
		klog.V(3).Infof("Token Expired")
		return true
	}
	klog.V(3).Infof("Token Not Expired")
	return false
}

/*
	RequestToken: Request a token by giving a private IPv4 addresses
	Input:
		host: Private IPv4 addresses
	Output:
		return a token string if no error, otherwise return the error
*/
func RequestToken(host string) (string, error) {
	tokenRequestURL := "http://" + host + "/identity/v3/auth/tokens"

	// TODO: Please don't hard code json data
	tokenJsonData := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","domain":{"id":"default"},"password":"secret"}}},"scope":{"project":{"name":"admin","domain":{"id":"default"}}}}}`

	// Make HTTP Request
	var tokenJsonDataBytes = []byte(tokenJsonData)
	req, _ := http.NewRequest("POST", tokenRequestURL, bytes.NewBuffer(tokenJsonDataBytes))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.V(3).Infof("HTTP Post Token Request Failed: %v", err)
		return "", err
	}
	klog.V(3).Infof("HTTP Post Request Succeeded")
	defer resp.Body.Close()

	// http.StatusCreated = 201
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("instance capacity has reached its limit")
	}

	// Token is stored in header
	respHeader := resp.Header
	klog.V(3).Infof("HTTP Post Token: %v", respHeader["X-Subject-Token"][0])

	return respHeader["X-Subject-Token"][0], nil
}
