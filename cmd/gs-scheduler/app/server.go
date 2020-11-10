package app

import (
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/apiserver"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/options"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/router"
)

// Run runt the service and other period monitor to make sure the consistency.
func Run(serverOptions *options.ServerRunOptions, stopCh <-chan struct{}) error {
	logger.Infof("Global Scheduler Running...")

	// init scheduler cache
	scheduler.InitSchedulerCache(stopCh)

	// init scheduler informer 
	scheduler.InitSchedulerInformer(stopCh)

	router.Register()
	hs, err := apiserver.NewHTTPServer()
	if err != nil {
		logger.Errorf("new http server failed!, err: %s", err.Error())
		return err
	}

	return hs.BlockingRun(stopCh)
}
