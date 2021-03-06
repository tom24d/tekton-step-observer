package step

import (
	"fmt"
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func GetSteps(taskrun *v1beta1.TaskRun) []v1beta1.Step {
	return taskrun.Status.TaskSpec.Steps
}

func GetStepStatuses(taskrun *v1beta1.TaskRun) []v1beta1.StepState {
	return taskrun.Status.Steps
}

func GetEventTime(state *v1beta1.StepState, eventType TektonPluginEventType) (*time.Time, error) {
	if eventType == CloudEventTypeStepStarted {
		if state.Running != nil && !state.Running.StartedAt.Time.IsZero() {
			return &state.Running.StartedAt.Time, nil
		} else if state.Terminated != nil && !state.Terminated.StartedAt.Time.IsZero() {
			return &state.Terminated.StartedAt.Time, nil
		}
		return nil, fmt.Errorf("no initialized ContainerState*.StartedAt")
	} else if eventType != CloudEventTypeStepSkipped {
		if state.Terminated != nil && !state.Terminated.FinishedAt.Time.IsZero() {
			return &state.Terminated.FinishedAt.Time, nil
		}
		return nil, fmt.Errorf("no initialized ContainerStateTerminated.FinishedAt")
	}
	// TODO consider time for skipped event
	return nil, fmt.Errorf("no time for skipped event")
}
