package controller

import (
	"context"
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/config"
	pipelineclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	taskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1beta1/taskrun"
	"github.com/tektoncd/pipeline/pkg/reconciler/events/cloudevent"

	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

const (
	controllerName = "step-observe-controller"
)

func NewController(ctx context.Context, cm configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)
	taskrunInformer := taskruninformer.Get(ctx)

	configStore := config.NewStore(logger.Named("step-observer"))
	configStore.WatchConfigs(cm)

	r := &Reconciler{
		taskRunLister: taskrunInformer.Lister(),
		pipelineClient: pipelineclient.Get(ctx),
		configStore: configStore,
		cloudEventClient: cloudevent.Get(ctx),
	}

	impl := controller.NewImpl(r, logger, controllerName)

	taskrunInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: impl.Enqueue,
		UpdateFunc: func(first, second interface{}) {
			// TODO remove this if the problem solved See: https://github.com/tom24d/step-observe-controller/issues/8
			impl.EnqueueAfter(second, 100 * time.Millisecond)
		},
	})

	return impl
}
