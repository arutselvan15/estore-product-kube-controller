// Package controllers controllers
package controllers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	cc "github.com/arutselvan15/estore-common/clients"
	"github.com/arutselvan15/estore-common/helper"
	pdtv1 "github.com/arutselvan15/estore-product-kube-client/pkg/apis/estore/v1"
	lc "github.com/arutselvan15/go-utils/logconstants"

	cfg "github.com/arutselvan15/estore-product-kube-controller/config"
)

// ProcessItemType process item type
type ProcessItemType func(*pdtv1.Product, cc.EstoreClientInterface, record.EventRecorder) error

// ProcessItem process item
func ProcessItem(pdt *pdtv1.Product, clients cc.EstoreClientInterface, recorder record.EventRecorder) error {
	log.SetObjectState(lc.Processing).SetStep(cfg.ProcessItem).SetStepState(lc.Start).Infof("process product %s start", pdt.Name)
	pdtCopy := pdt.DeepCopy()

	// examine DeletionTimestamp to determine if object is under deletion
	if pdtCopy.ObjectMeta.DeletionTimestamp.IsZero() {
		if err := update(pdtCopy, clients, recorder); err != nil {
			handleError(pdtCopy, err, recorder)
			return err
		}

		pdtCopy.Status.CurrentStatus.Phase = pdtv1.ProductAvailable
		pdtCopy.Status.CurrentStatus.LastUpdateTime = metav1.Now()
		pdtCopy.Status.LastOperation.LastUpdateTime = metav1.Now()

		if _, err := clients.GetProductClient().EstoreV1().Products(pdtCopy.Namespace).UpdateStatus(pdtCopy); err != nil {
			handleError(pdtCopy, err, recorder)
			return err
		}

		recorder.Event(pdtCopy, corev1.EventTypeNormal, "Phase", "Available")
	} else if helper.ContainsString(pdtCopy.ObjectMeta.Finalizers, cfg.ProductOperatorFinalizer) {
		// The object is being deleted
		// our finalizer is present, so lets handle any external dependency
		if err := delete(pdtCopy, clients, recorder); err != nil {
			// fail to delete the external dependency here, return with error so that it can be retried
			handleError(pdtCopy, err, recorder)
			return err
		}

		// remove our finalizer from the list and update it.
		pdtCopy.ObjectMeta.Finalizers = helper.RemoveString(pdtCopy.ObjectMeta.Finalizers, cfg.ProductOperatorFinalizer)
		pdtCopy.Status.CurrentStatus.LastUpdateTime = metav1.Now()
		pdtCopy.Status.LastOperation.LastUpdateTime = metav1.Now()

		if _, err := clients.GetProductClient().EstoreV1().Products(pdtCopy.Namespace).Update(pdtCopy); err != nil {
			handleError(pdtCopy, err, recorder)
			return err
		}

		recorder.Event(pdtCopy, corev1.EventTypeNormal, "Phase", "Deleted")
	}

	log.SetObjectState(lc.Successful).SetStepState(lc.Complete).Infof("process product %s completed successfully", pdt.Name)

	return nil
}

func update(pdtCopy *pdtv1.Product, clients cc.EstoreClientInterface, recorder record.EventRecorder) error {
	log.SetStepState(lc.Processing).Debugf("processing pdt %s", pdtCopy.Name)

	return nil
}

func delete(pdtCopy *pdtv1.Product, clients cc.EstoreClientInterface, recorder record.EventRecorder) error {
	log.SetStepState(lc.Processing).Debugf("processing pdt %s", pdtCopy.Name)

	return nil
}

func handleError(pdtCopy *pdtv1.Product, err error, recorder record.EventRecorder) {
	recorder.Event(pdtCopy, corev1.EventTypeWarning, "Reason", err.Error())
	recorder.Event(pdtCopy, corev1.EventTypeWarning, "Phase", "Unavailable")
	log.SetStepState(lc.Error).Error(err.Error())
	log.SetStepState(lc.Retry).Debugf("process product %s failed, re-queued for retry", pdtCopy.Name)
}
