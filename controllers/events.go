// Package controllers controllers
package controllers

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	pdtv1 "github.com/arutselvan15/estore-product-kube-client/pkg/apis/estore/v1"
	lc "github.com/arutselvan15/go-utils/logconstants"
)

func onAdd(pdt *pdtv1.Product, recorder record.EventRecorder) string {
	// checks for status if it is already available then dont add its resync
	if pdt.Status.CurrentStatus.Phase == pdtv1.ProductAvailable {
		return ""
	}

	key, err := cache.MetaNamespaceKeyFunc(pdt)
	if err != nil {
		return ""
	}

	recorder.Event(pdt, corev1.EventTypeNormal, "Event", "Create")
	log.SetOperation(lc.Create).SetObjectName(pdt.Name).SetObjectState(lc.Received).SetStep("").SetStepState("").LogAuditObject(pdt)

	return key
}

func onUpdate(oldPdt, pdt *pdtv1.Product, recorder record.EventRecorder) string {
	// checks for status if it is already available then dont add its resync
	if pdt.Status.CurrentStatus.Phase == pdtv1.ProductAvailable {
		return ""
	}

	key, err := cache.MetaNamespaceKeyFunc(pdt)
	if err != nil {
		return ""
	}

	recorder.Event(pdt, corev1.EventTypeNormal, "Event", "Update")
	log.SetOperation(lc.Update).SetObjectName(pdt.Name).SetObjectState(lc.Received).SetStep("").SetStepState("").LogAuditObject(oldPdt, pdt)

	return key
}

func onDelete(pdt *pdtv1.Product, recorder record.EventRecorder) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(pdt)
	if err != nil {
		return ""
	}

	recorder.Event(pdt, corev1.EventTypeNormal, "Event", "Delete")
	log.SetOperation(lc.Delete).SetObjectName(pdt.Name).SetObjectState(lc.Received).SetStep("").SetStepState("").LogAuditObject(pdt)

	return key
}
