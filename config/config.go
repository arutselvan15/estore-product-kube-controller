// Package config config
package config

import "time"

const (
	// ResourceName resource name
	ResourceName = "product"
	// Component component name
	Component = "controller"
	// WorkQueueName work queue name
	WorkQueueName = "product"
	// ExitErrorCode exit code
	ExitErrorCode = 1
	// ProcessItem process item
	ProcessItem = "processItem"
	// ProductOperatorFinalizer finalizers
	ProductOperatorFinalizer = "operator.finalizers.product.estore.com"
)

var (
	// ResyncDuration resync duration in minutes
	ResyncDuration = 15 * time.Minute
	// WorkerCount worker count
	WorkerCount = 1
)
