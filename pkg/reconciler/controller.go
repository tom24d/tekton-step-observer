package reconciler

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	"github.com/tektoncd/pipeline/pkg/apis/config"
	pipelineclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	taskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1beta1/taskrun"
	cloudeventclient "github.com/tektoncd/pipeline/pkg/reconciler/events/cloudevent"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

const (
	controllerName = "step-observer"
)

func NewController(ctx context.Context, cm configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)
	taskrunInformer := taskruninformer.Get(ctx)

	configStore := config.NewStore(logger.Named("step-observer"))
	configStore.WatchConfigs(cm)

	r := &Reconciler{
		LeaderAwareFuncs: reconciler.LeaderAwareFuncs{
			PromoteFunc: func(bkt reconciler.Bucket, enq func(reconciler.Bucket, types.NamespacedName)) error {
				all, err := taskrunInformer.Lister().List(labels.Everything())
				if err != nil {
					return err
				}
				for _, elt := range all {
					// TODO: Consider letting users specify a filter in options.
					enq(bkt, types.NamespacedName{
						Namespace: elt.GetNamespace(),
						Name:      elt.GetName(),
					})
				}
				return nil
			},
		},

		taskRunLister:    taskrunInformer.Lister(),
		pipelineClient:   pipelineclient.Get(ctx),
		kubeClientSet:    kubeclient.Get(ctx),
		configStore:      configStore,
		cloudEventClient: cloudeventclient.Get(ctx),
	}

	impl := controller.NewImpl(r, logger, controllerName)

	taskrunInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: impl.Enqueue,
		UpdateFunc: func(first, second interface{}) {
			oldObj, ok := first.(kmeta.Accessor)
			if !ok {
				return
			}
			newObj, ok := second.(kmeta.Accessor)
			if !ok {
				return
			}
			if oldObj.GetResourceVersion() != newObj.GetResourceVersion() {
				// TODO remove this if the problem solved. See: https://github.com/tom24d/step-observe-controller/issues/8
				impl.EnqueueAfter(second, 100*time.Millisecond)
			}
		},
	})

	return impl
}
