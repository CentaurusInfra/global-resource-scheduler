/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of Arktos - file modified.

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

package apiserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	_ "k8s.io/kubernetes/globalscheduler/pkg/scheduler"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/config"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
)

const (
	defaultKeepAlivePeriod = 3 * time.Minute
)

// HTTPServer contains a http server abnd
type HTTPServer struct {
	httpServer *http.Server
	listener   net.Listener
}

func getServerAddress() string {
	hostIP := config.DefaultString("address", "0.0.0.0")
	if hostIP == "" {
		logger.Errorf("server IP address not configured.")
		os.Exit(1)
	}

	port := config.DefaultInt("port", 8443)

	return fmt.Sprintf("%s:%d", hostIP, port)
}

// NewHTTPServer construct a new http server with ssl certificate
func NewHTTPServer() (*HTTPServer, error) {
	hs := &HTTPServer{}

	httpAddr := getServerAddress()
	l, err := net.Listen("tcp", httpAddr)
	if err != nil {
		logger.Errorf("failed to http(https) listen: %s err: %s", httpAddr, err.Error())
		return nil, err
	}
	hs.listener = l
	hs.httpServer = &http.Server{
		Addr:           l.Addr().String(),
		MaxHeaderBytes: 1 << 20,
	}

	return hs, nil
}

// BlockingRun make server running with blocking and shutdown with certain time timeout
func (hs *HTTPServer) BlockingRun(stopCh <-chan struct{}) error {
	return RunServer(hs.httpServer, hs.listener, 60, stopCh)
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
//
// Copied from Go 1.7.2 net/http/server.go
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	err = tc.SetKeepAlive(true)
	if err != nil {
		logger.Infof("Meet error when setting KeepAlive as true: %s", err.Error())
	}
	err = tc.SetKeepAlivePeriod(defaultKeepAlivePeriod)
	if err != nil {
		logger.Infof("Meet error when setting KeepAlivePeriod (%s): %s", defaultKeepAlivePeriod, err.Error())
	}
	return tc, nil
}

// RunServer run server gracefully
func RunServer(
	server *http.Server,
	ln net.Listener,
	shutDownTimeout time.Duration,
	stopCh <-chan struct{}) error {
	if ln == nil {
		return fmt.Errorf("listener must not be nil")
	}

	// Shutdown server gracefully.
	go func() {
		<-stopCh
		logger.Infof("shutdown Server gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), shutDownTimeout)
		err := server.Shutdown(ctx)
		if err != nil {
			logger.Errorf("shutdown Server failed, err: %s.", err.Error())
		}
		cancel()
	}()

	var listener net.Listener
	listener = tcpKeepAliveListener{ln.(*net.TCPListener)}

	if server.TLSConfig != nil {
		listener = tls.NewListener(listener, server.TLSConfig)
	}

	err := server.Serve(listener)
	if err != nil {
		logger.Errorf("Server runs failed, err: %s.", err.Error())
	}

	msg := fmt.Sprintf("Stopped listening on %s", ln.Addr().String())
	select {
	case _, ok := <-stopCh:
		if !ok {
			return nil
		}
		fmt.Println("continue")
	default:
		errMsg := fmt.Sprintf("%s due to error: %v", msg, err.Error())
		fmt.Println(errMsg)
		os.Exit(1)
	}

	return nil
}
