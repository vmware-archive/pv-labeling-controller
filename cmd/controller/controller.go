package main

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconcilePersistentVolume struct {
	client client.Client
	log    logr.Logger

	coreClient      kubernetes.Interface
	labelKeysToSync []string
}

var _ reconcile.Reconciler = &reconcilePersistentVolume{}

func (r *reconcilePersistentVolume) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("request", request)

	pv := &corev1.PersistentVolume{}

	err := r.client.Get(context.TODO(), request.NamespacedName, pv)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error(nil, "Could not find PersistentVolume")
			return reconcile.Result{}, nil
		}

		log.Error(err, "Could not fetch PersistentVolume")
		return reconcile.Result{}, err
	}

	if pv.Spec.ClaimRef != nil {
		pvc, err := r.coreClient.CoreV1().PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			log.Error(nil, "Could not find PersistentVolumeClaim")
			return reconcile.Result{}, nil
		}

		if pv.Labels == nil {
			pv.Labels = map[string]string{}
		}

		// Create, update or delete
		for _, lblKey := range r.labelKeysToSync {
			if val, found := pvc.Labels[lblKey]; found {
				pv.Labels[lblKey] = val
			} else {
				delete(pv.Labels, lblKey)
			}
		}
	} else {
		// Delete all label keys
		for _, lblKey := range r.labelKeysToSync {
			delete(pv.Labels, lblKey)
		}
	}

	err = r.client.Update(context.TODO(), pv)
	if err != nil {
		log.Error(err, "Could not write PersistentVolume")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

/*

kind: PersistentVolume
spec:
  claimRef:
    apiVersion: v1
    kind: PersistentVolumeClaim
    name: cassandra-data-cassandra-0
    namespace: default
    resourceVersion: "58113337"
    uid: 8264ba89-acdc-11e9-bf6e-42010a800002

*/
