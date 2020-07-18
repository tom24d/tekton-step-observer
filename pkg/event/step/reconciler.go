package step

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/types"

	"knative.dev/pkg/logging"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	v1beta1client "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1beta1"

	"github.com/tom24d/step-observe-controller/pkg/event/step/resources"
)

const (
	annotationKey = "tom24d.plugin.step-observer/result"
)

// ReconcileStepEvent actually reconciles taskrun to determine whether it should emit CloudEvent.
// TODO maybe it has to watch Pod also and GET about step container spec&status directly.
func ReconcileStepEvent(ctx context.Context, taskrun *v1beta1.TaskRun, client v1beta1client.TaskRunInterface) error {
	logger := logging.FromContext(ctx)
	if taskrun.Status.TaskSpec == nil ||
		len(taskrun.Status.TaskSpec.Steps) < 1 {
		logger.Infof("step event emission skipped as no step in taskrun: %s", taskrun.Name)
		return nil
	}

	sent, err := initializeAnnotation(taskrun)
	if err != nil {
		logger.Fatalf("failed to initialize annotation: %v", err)
		return nil
	}


	for i, step := range taskrun.Status.Steps {
		// no emission if already sent

		// start
		if step.ContainerState.Running != nil {
			// (Running) emit&mark started if i=0
			if i == 0 {
				ensureEventEmitted(ctx, sent, step.Name, resources.CloudEventTypeStepStarted)
			}
		}
		// emit&mark started if i!=0 && i-1 step marked as succeeded
		if i != 0 && sent.IsMarked(taskrun.Status.Steps[i-1].Name, resources.CloudEventTypeStepSucceeded) {
			ensureEventEmitted(ctx, sent, step.Name, resources.CloudEventTypeStepStarted)
		}

		if step.ContainerState.Terminated != nil {
			// skipped
			// (Terminated) emit&mark skipped if i!=0 && i-1 step marked as failed||skipped,
			// then continue to avoid being considered as failure
			if i != 0 &&
				(sent.IsMarked(taskrun.Status.Steps[i-1].Name, resources.CloudEventTypeStepFailed) ||
					sent.IsMarked(taskrun.Status.Steps[i-1].Name, resources.CloudEventTypeStepSkipped)) {
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

		// otherwise, no idea
	}

	// TODO PATCH for updated annotation at last
	data, _ := sent.MarshalString()
	tr := v1beta1.TaskRun{}
	tr.Annotations = make(map[string]string)
	tr.Annotations[annotationKey] = data
	// TODO error handling
	patch, _ := json.Marshal(tr)
	logger.Infof("updated annotationKey: %s", sent)
	_, err = client.Patch(taskrun.Name, types.MergePatchType, patch)
	if err != nil {
		logger.Errorf("patch error: %v", err)
	}
	return nil
}

func ensureEventEmitted(
	ctx context.Context,
	annotation *resources.CloudEventSent,
	stepName string,
	eventType resources.CloudEventType) *resources.CloudEventSent {

	logger := logging.FromContext(ctx)
	if !annotation.IsMarked(stepName, eventType) {
		logger.Infof("EVENT EMISSION for step: %s, type: %v", stepName, eventType)
		annotation, err := annotation.MarkEvent(ctx, stepName, eventType)
		if err != nil {
			logger.Fatalf("error occured at step: %v", err)
		}
		return annotation
	}
	return annotation
}

func initializeAnnotation(run *v1beta1.TaskRun) (*resources.CloudEventSent, error) {
	annotation := &resources.CloudEventSent{}
	if val, ok := run.Annotations[annotationKey]; ok {
		err := resources.UnmarshalString(val, annotation)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotation: %v", err)
		}
		return annotation, nil
	}

	steps := run.Status.TaskSpec.Steps
	for _, step := range steps {
		r := resources.Reported{StepName: step.Name, ReportedType: make([]resources.CloudEventType, 0, 2)}
		annotation.Steps = append(annotation.Steps, r)
	}

	return annotation, nil
}
