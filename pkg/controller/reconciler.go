package controller

import (
	"context"

	"knative.dev/pkg/logging"

	listers "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1beta1"
)

type Reconciler struct {
	taskRunLister listers.TaskRunLister
}

func (r Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)
	logger.Infof("Hello reconciler. key: %v", key)

	return nil
}
