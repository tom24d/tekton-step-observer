package controller

import (
	"context"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	pipelineclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	taskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1beta1/taskrun"
)

func NewController(ctx context.Context, cm configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)
	taskrunInformer := taskruninformer.Get(ctx)

	r := &Reconciler{
		taskRunLister: taskrunInformer.Lister(),
		pipelineClient: pipelineclient.Get(ctx),
	}

	impl := controller.NewImpl(r, logger, "my-controller")

	taskrunInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	return impl
}
