package step

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/logging"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	v1beta1client "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1beta1"

	"github.com/tom24d/step-observe-controller/pkg/events/step/resources"
)

const (
	annotationKey = "tom24d.plugin.step-observer/result"
)

// ReconcileStepEvent is entry point to reconcile taskrun to determine whether it should emit CloudEvent.
func ReconcileStepEvent(ctx context.Context, taskrun *v1beta1.TaskRun, client v1beta1client.TaskRunInterface) error {
	logger := logging.FromContext(ctx)
	if taskrun.Status.TaskSpec == nil ||
		len(GetSteps(taskrun)) < 1 {
		logger.Infof("step events emission skipped as no step in taskrun: %s", taskrun.Name)
		return nil
	}

	sent, err := initializeAnnotation(taskrun)
	if err != nil {
		logger.Fatalf("failed to initialize annotation: %v", err)
		return nil
	}
	logger.Infof("RECEIVED annotation: %s, annotation in the run: %v", sent, taskrun.Annotations)


	for i, step := range GetStepStatuses(taskrun) {
		// start
		if step.ContainerState.Running != nil {
			// (Running) emit&mark started if i=0
			if i == 0 && !taskrun.Status.StartTime.IsZero(){
				ensureEventEmitted(ctx, sent, step.Name, resources.CloudEventTypeStepStarted)
			}
		}
		// emit&mark started if i!=0 && i-1 step marked as succeeded
		if i != 0 && sent.IsMarked(GetStepStatuses(taskrun)[i-1].Name, resources.CloudEventTypeStepSucceeded) {
			ensureEventEmitted(ctx, sent, step.Name, resources.CloudEventTypeStepStarted)
		}

		if step.ContainerState.Terminated != nil {
			// skipped
			// (Terminated) emit&mark skipped if i!=0 && i-1 step marked as failed||skipped,
			// then continue to avoid being considered as failure
			if i != 0 &&
					(sent.IsMarked(GetStepStatuses(taskrun)[i-1].Name, resources.CloudEventTypeStepFailed) ||
					sent.IsMarked(GetStepStatuses(taskrun)[i-1].Name, resources.CloudEventTypeStepSkipped)) {

				ensureEventEmitted(ctx, sent, step.Name, resources.CloudEventTypeStepSkipped)
				continue
			}
			// succeeded
			// (Terminated) emit&mark succeeded if exit code is 0
			if step.Terminated.ExitCode == 0 {
				ensureEventEmitted(ctx, sent, step.Name, resources.CloudEventTypeStepSucceeded)
			} else {
				// failure
				// (Terminated) emit&mark failed if exit code is not 0
				ensureEventEmitted(ctx, sent, step.Name, resources.CloudEventTypeStepFailed)
			}

		}
	}

	// TODO proper reconcileErr handling
	patch, err := getPatch(sent)
	if err != nil {
		return err
	}
	logger.Infof("PATCH: %s", patch)
	_, err = client.Patch(taskrun.Name, types.MergePatchType, patch)
	if err != nil {
		logger.Errorf("failed to PATCH taskrun: %v", err)
	}
	return nil
}

func ensureEventEmitted(ctx context.Context, annotation *resources.EmissionStatuses, stepName string, eventType resources.TektonPluginEventType) {
	logger := logging.FromContext(ctx)

	emissionStatus, err := annotation.GetStatus(stepName)
	if err != nil {
		logger.Fatalf("status of step: %s not found", stepName)
		return
	}

	// no emission if already marked as sent
	if !emissionStatus.IsMarked(eventType) {
		//logger.Infof("EVENT annotation: %v", annotation)
		err := emissionStatus.MarkEvent(eventType)
		logger.Infof("EVENT EMISSION step: %s, type: %v, annotation: %v", stepName,  eventType, annotation)
		if err != nil {
			logger.Fatalf("error occured at step-%s: %v", stepName, err)
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

	for _, step := range GetSteps(run) {
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
