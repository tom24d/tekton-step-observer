package reconciler

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"github.com/tom24d/step-observe-controller/pkg/events/step/resources"
)

const (
	annotationKey = "tekton.plugin.step-observer/result"
)

type Reconciler struct {
	taskRunLister    listers.TaskRunLister
	pipelineClient   clientset.Interface
	kubeClientSet kubernetes.Interface
	configStore      reconciler.ConfigStore
	cloudEventClient cloudevent.CEClient
}

// Reconcile reconciles taskrun resource for emitting CloudEvents.
// This reconciler does not change any spec/status. All info is stored as json in the metadata.annotation
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
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
	if taskrun.Status.TaskSpec == nil ||
		len(stepresource.GetSteps(taskrun)) < 1 {
		logger.Infof("step events emission skipped as no step in taskrun: %s", taskrun.Name)
		return nil
	}

	sent, err := initializeAnnotation(taskrun)
	if err != nil {
		logger.Fatalf("failed to initialize annotation: %v", err)
		return nil
	}
	logger.Infof("RECEIVED annotation: %s, annotation in the run: %v", sent, taskrun.Annotations)

	for i, step := range stepresource.GetStepStatuses(taskrun) {
		// start
		if step.ContainerState.Running != nil {
			// (Running) emit&mark started if i=0
			if i == 0 && !taskrun.Status.StartTime.IsZero() {
				r.ensureEventEmitted(ctx, sent, resources.CloudEventTypeStepStarted, taskrun, i)
			}
		}
		// emit&mark started if i!=0 && i-1 step marked as succeeded
		if i != 0 && sent.IsMarked(stepresource.GetStepStatuses(taskrun)[i-1].Name, resources.CloudEventTypeStepSucceeded) {
			r.ensureEventEmitted(ctx, sent, resources.CloudEventTypeStepStarted, taskrun, i)
		}

		if step.ContainerState.Terminated != nil {
			// skipped
			// (Terminated) emit&mark skipped if i!=0 && i-1 step marked as failed||skipped,
			// then continue to avoid being considered as failure
			if i != 0 &&
				(sent.IsMarked(stepresource.GetStepStatuses(taskrun)[i-1].Name, resources.CloudEventTypeStepFailed) ||
					sent.IsMarked(stepresource.GetStepStatuses(taskrun)[i-1].Name, resources.CloudEventTypeStepSkipped)) {

				r.ensureEventEmitted(ctx, sent, resources.CloudEventTypeStepSkipped, taskrun, i)
				continue
			}
			// succeeded
			// (Terminated) emit&mark succeeded if exit code is 0
			if step.Terminated.ExitCode == 0 {
				r.ensureEventEmitted(ctx, sent, resources.CloudEventTypeStepSucceeded, taskrun, i)
			} else {
				// failure
				// (Terminated) emit&mark failed if exit code is not 0
				r.ensureEventEmitted(ctx, sent, resources.CloudEventTypeStepFailed, taskrun, i)
			}

		}
	}

	// TODO proper reconcileErr handling
	patch, err := getPatch(sent)
	if err != nil {
		return err
	}
	logger.Infof("PATCH: %s", patch)
	_, err = r.pipelineClient.TektonV1beta1().TaskRuns(taskrun.Namespace).Patch(taskrun.Name, types.MergePatchType, patch)
	if err != nil {
		logger.Errorf("failed to PATCH taskrun: %v", err)
	}
	return nil
}

func (r *Reconciler) ensureEventEmitted(
	ctx context.Context, annotation *resources.EmissionStatuses, eventType resources.TektonPluginEventType,
	run *v1beta1.TaskRun, index int,
) {
	logger := logging.FromContext(ctx)

	name := stepresource.GetSteps(run)[index].Name
	emissionStatus, err := annotation.GetStatus(name)
	if err != nil {
		logger.Fatalf("status of step: %s not found", name)
		return
	}
	s1 := stepresource.GetSteps(run)
	s2 := stepresource.GetStepStatuses(run)

	// no emission if already marked as sent
	if !emissionStatus.IsMarked(eventType) {
		p, err := r.getPod(run)
		if err != nil {
			return
		}
		log, err := r.getStepLog(run, index)
		if err != nil {
			return
		}
		data := resources.TektonStepCloudEvent{
			Detail: &s1[index],
			State:  &s2[index],
			Pod:    p,
			Log:    log,
		}
		go data.Emit(ctx, eventType)
		err = emissionStatus.MarkEvent(eventType)
		logger.Infof("EVENT EMISSION step: %s, type: %v, annotation: %v", name, eventType, annotation)
		if err != nil {
			logger.Fatalf("error occured at step-%s: %v", name, err)
		}
		return
	}
}

func initializeAnnotation(run *v1beta1.TaskRun) (*resources.EmissionStatuses, error) {
	annotation := &resources.EmissionStatuses{}
	if val, ok := run.Annotations[annotationKey]; ok {
		err := resources.UnmarshalString(val, annotation)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotation: %v", err)
		}
		return annotation, nil
	}

	for _, step := range stepresource.GetSteps(run) {
		r := resources.EmissionStatus{Name: step.Name, Emitted: make([]resources.TektonPluginEventType, 0, 2)}
		annotation.Statuses = append(annotation.Statuses, r)
	}

	return annotation, nil
}

func getPatch(state *resources.EmissionStatuses) ([]byte, error) {
	data, err := state.MarshalString()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal EmissionStatuses object: %v", err)
	}
	tr := v1beta1.TaskRun{}
	tr.Annotations = make(map[string]string)
	tr.Annotations[annotationKey] = data
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

func (r *Reconciler) getPod(run *v1beta1.TaskRun) (*corev1.Pod, error) {
	ref := run.GetBuildPodRef()
	return r.kubeClientSet.CoreV1().Pods(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
}
