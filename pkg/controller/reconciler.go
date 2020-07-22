package controller

import (
	"context"
	"go.uber.org/zap"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	clientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	listers "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/reconciler/events/cloudevent"

	stepevent "github.com/tom24d/step-observe-controller/pkg/events/step"
)

type Reconciler struct {
	taskRunLister  listers.TaskRunLister
	pipelineClient clientset.Interface
	configStore reconciler.ConfigStore
	cloudEventClient cloudevent.CEClient
}

// Reconcile reconciles taskrun resource for emitting CloudEvents.
// This reconciler does not change any spec/status. All info is stored as json in the metadata.annotation
func (r Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)
	ctx = cloudevent.ToContext(ctx, r.cloudEventClient)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("invalid resource key: %s", zap.Any("key", key))
	}

	taskrun, ok := r.taskRunLister.TaskRuns(namespace).Get(name)
	if apierrors.IsNotFound(ok) {
		logger.Error("could not find taskrun", zap.Any("key", key))
		return nil
	} else if !taskrun.DeletionTimestamp.IsZero() {
		return nil
	} else if ok != nil {
		return err
	}

	taskrun = taskrun.DeepCopy()

	reconcileErr := stepevent.ReconcileStepEvent(ctx, taskrun, r.pipelineClient.TektonV1beta1().TaskRuns(namespace))
	if reconcileErr != nil {
		logger.Warn("Error reconciling TaskRun for step observer:", zap.Error(reconcileErr))
	} else {
		logger.Debug("TaskRun reconciled for step observer")
		// consider emitting k8s events to record failures of CloudEvent emission.
	}

	return nil
}
