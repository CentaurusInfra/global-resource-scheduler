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

package utils

import (
	"crypto/tls"
	"io"
	"net/http"
	"time"
)

// GetTransport get Transport
func GetTransport(isClientCARequired bool) *http.Transport {
	tr, err := initTransport(isClientCARequired)
	if err != nil {
	}
	return tr
}

// initTransport initialize http
func initTransport(isClientCARequired bool) (*http.Transport, error) {
	tr := &http.Transport{DisableKeepAlives: true}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return tr, nil
}

// SendHTTPRequest send http request
func SendHTTPRequest(method string, serverAddr string, headers map[string][]string,
	body io.Reader, isClientCARequired bool) (r *http.Response, e error) {

	client := &http.Client{Transport: GetTransport(isClientCARequired), Timeout: time.Duration(2 * time.Minute)}

	req, err := http.NewRequest(method, serverAddr, body)
	if err != nil {
		return nil, err
	}
	req.Close = true

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
