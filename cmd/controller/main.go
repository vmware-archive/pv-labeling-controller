package main

// Based on https://github.com/kubernetes-sigs/controller-runtime/blob/8f633b179e1c704a6e40440b528252f147a3362a/examples/builtins/main.go

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("pv-labeling-controller")

func main() {
	logf.SetLogger(zap.Logger(false))
	entryLog := log.WithName("entrypoint")

	entryLog.Info("setting up manager")

	restConfig := config.GetConfigOrDie()

	mgr, err := manager.New(restConfig, manager.Options{})
	if err != nil {
		entryLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	entryLog.Info("Setting up controller")

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		entryLog.Error(err, "building core client")
		os.Exit(1)
	}

	reconciler := &reconcilePersistentVolume{
		client:          mgr.GetClient(),
		coreClient:      coreClient,
		labelKeysToSync: []string{"kapp.k14s.io/app", "kapp.k14s.io/association"},
		log:             log.WithName("reconciler"),
	}

	c, err := controller.New("pv-labeling-controller", mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		entryLog.Error(err, "unable to set up individual controller")
		os.Exit(1)
	}

	if err := c.Watch(&source.Kind{Type: &corev1.PersistentVolume{}}, &handler.EnqueueRequestForObject{}); err != nil {
		entryLog.Error(err, "unable to watch PeristentVolume")
		os.Exit(1)
	}

	entryLog.Info("starting manager")

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
