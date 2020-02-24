// Package log provides logger
package log

import (
	"sync"

	cl "github.com/arutselvan15/estore-common/log"
	gLog "github.com/arutselvan15/go-utils/log"

	cfg "github.com/arutselvan15/estore-product-kube-controller/config"
)

var (
	logInstance gLog.CommonLog
	once        sync.Once
)

// GetLogger the Log object
func GetLogger() gLog.CommonLog {
	once.Do(func() {
		logInstance = cl.GetLogger(cfg.ResourceName).SetComponent(cfg.Component)
	})

	return logInstance
}
