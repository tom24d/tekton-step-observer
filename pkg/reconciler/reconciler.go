package reconciler

import (
	"context"
	"encoding/json"
	"fmt"

	multierr "github.com/hashicorp/go-multierror"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	clientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	listers "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/reconciler/events/cloudevent"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	stepresource "github.com/tom24d/step-observe-controller/pkg/events/step"
)

const (
	AnnotationKey = "tekton.plugin.step-observer/result"
)

type Reconciler struct {
	taskRunLister    listers.TaskRunLister
	pipelineClient   clientset.Interface
	kubeClientSet    kubernetes.Interface
	configStore      reconciler.ConfigStore
	cloudEventClient cloudevent.CEClient
}

// Reconcile reconciles taskrun resource for emitting CloudEvents.
// This reconciler does not change any spec/status. All info is stored as json in the metadata.annotation
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)
	ctx = r.configStore.ToContext(ctx)
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

	reconcileErr := r.reconcile(ctx, taskrun)
	if reconcileErr != nil {
		logger.Warn("Error reconciling TaskRun for step observer:", zap.Error(reconcileErr))
	} else {
		logger.Debug("TaskRun reconciled for step observer")
		// consider emitting k8s events to record failures of CloudEvent emission.
	}

	return nil
}

// ReconcileStepEvent is entry point to reconcile taskrun to determine whether it should emit CloudEvent.
func (r *Reconciler) reconcile(ctx context.Context, taskrun *v1beta1.TaskRun) error {
	logger := logging.FromContext(ctx)
	if taskrun.Status.TaskSpec == nil || len(stepresource.GetSteps(taskrun)) < 1 {
		logger.Infof("step events emission skipped as no step in the taskrun: %s", taskrun.Name)
		return nil
	}

	sent, err := initializeAnnotation(taskrun)
	if err != nil {
		logger.Fatalf("failed to initialize annotation: %v", err)
		return err
	}

	if e := r.reconcileSteps(ctx, sent, taskrun); e != nil {
		return e
	}

	// TODO proper reconcileErr handling
	patch, err := getPatch(sent)
	if err != nil {
		return err
	}
	_, err = r.pipelineClient.TektonV1beta1().TaskRuns(taskrun.Namespace).Patch(taskrun.Name, types.MergePatchType, patch)
	if err != nil {
		logger.Errorf("failed to PATCH taskrun: %v", err)
	}
	return nil
}

func (r *Reconciler) reconcileSteps(ctx context.Context, sent *stepresource.EmissionStatuses, taskrun *v1beta1.TaskRun) error {
	var errs *multierr.Error
	for i, step := range stepresource.GetStepStatuses(taskrun) {
		// start
		if step.ContainerState.Running != nil {
			// (Running) emit&mark started if i=0
			if i == 0 && !taskrun.Status.StartTime.IsZero() {
				errs = multierr.Append(r.ensureEventEmitted(ctx, sent, stepresource.CloudEventTypeStepStarted, taskrun, i))
			}
		}
		// emit&mark started if i!=0 && i-1 step marked as succeeded
		if i != 0 && sent.IsMarked(stepresource.GetStepStatuses(taskrun)[i-1].Name, stepresource.CloudEventTypeStepSucceeded) {
			errs = multierr.Append(r.ensureEventEmitted(ctx, sent, stepresource.CloudEventTypeStepStarted, taskrun, i))
		}

		if step.ContainerState.Terminated != nil {
			// skipped
			// (Terminated) emit&mark skipped if i!=0 && i-1 step marked as failed||skipped,
			// then continue to avoid being considered as failure
			if i != 0 &&
				(sent.IsMarked(stepresource.GetStepStatuses(taskrun)[i-1].Name, stepresource.CloudEventTypeStepFailed) ||
					sent.IsMarked(stepresource.GetStepStatuses(taskrun)[i-1].Name, stepresource.CloudEventTypeStepSkipped)) {

				errs = multierr.Append(r.ensureEventEmitted(ctx, sent, stepresource.CloudEventTypeStepSkipped, taskrun, i))
				continue
			}
			// succeeded
			// (Terminated) emit&mark succeeded if exit code is 0
			if step.Terminated.ExitCode == 0 {
				errs = multierr.Append(r.ensureEventEmitted(ctx, sent, stepresource.CloudEventTypeStepSucceeded, taskrun, i))
			} else {
				// failure
				// (Terminated) emit&mark failed if exit code is not 0
				errs = multierr.Append(r.ensureEventEmitted(ctx, sent, stepresource.CloudEventTypeStepFailed, taskrun, i))
			}
		}
	}

	return errs.ErrorOrNil()
}

func (r *Reconciler) ensureEventEmitted(
	ctx context.Context, annotation *stepresource.EmissionStatuses, eventType stepresource.TektonPluginEventType,
	run *v1beta1.TaskRun, index int,
) error {
	name := stepresource.GetSteps(run)[index].Name
	emissionStatus, err := annotation.GetStatus(name)
	if err != nil {
		return err
	}
	s1 := stepresource.GetSteps(run)
	s2 := stepresource.GetStepStatuses(run)

	// no emission if already marked as sent
	if !emissionStatus.IsMarked(eventType) {
		log, err := r.getStepLog(run, index)
		if err != nil {
			return err
		}
		data := stepresource.TektonStepCloudEvent{
			Step:      &s1[index],
			StepState: &s2[index],
			PodRef: &corev1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Pod",
				Name:       run.Status.PodName,
				Namespace:  run.Namespace,
			},
			Log: log,
		}
		go data.Emit(ctx, eventType)
		err = emissionStatus.MarkEvent(eventType)
		if err != nil {
			return err
		}
	}
	return nil
}

func initializeAnnotation(run *v1beta1.TaskRun) (*stepresource.EmissionStatuses, error) {
	annotation := &stepresource.EmissionStatuses{}
	if val, ok := run.Annotations[AnnotationKey]; ok {
		err := json.Unmarshal([]byte(val), annotation)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotation: %v", err)
		}
		return annotation, nil
	}

	for _, step := range stepresource.GetStepStatuses(run) {
		r := stepresource.EmissionStatus{
			Name:    step.Name,
			Emitted: make([]stepresource.TektonPluginEventType, 0, 2),
		}
		annotation.Statuses = append(annotation.Statuses, r)
	}

	return annotation, nil
}

func getPatch(state *stepresource.EmissionStatuses) ([]byte, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal EmissionStatuses object: %v", err)
	}
	tr := v1beta1.TaskRun{}
	tr.Annotations = make(map[string]string)
	tr.Annotations[AnnotationKey] = string(data)
	return json.Marshal(tr)
}

func (r *Reconciler) getStepLog(run *v1beta1.TaskRun, i int) (string, error) {
	stepState := stepresource.GetStepStatuses(run)
	podName := run.Status.PodName
	containerName := stepState[i].ContainerName
	logOpt := &corev1.PodLogOptions{
		Container: containerName,
	}

	result := r.kubeClientSet.CoreV1().Pods(run.Namespace).GetLogs(podName, logOpt).Do()
	log, err := result.Raw()
	return fmt.Sprintf("%s", log), err
}
