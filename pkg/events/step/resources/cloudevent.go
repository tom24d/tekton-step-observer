package resources

import (
	corev1 "k8s.io/api/core/v1"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

type TektonPluginEventType string

const (
	CloudEventTypeStepStarted   TektonPluginEventType = "dev.tekton.events.plugin.step.started.v1"
	CloudEventTypeStepFailed    TektonPluginEventType = "dev.tekton.events.plugin.step.failed.v1"
	CloudEventTypeStepSucceeded TektonPluginEventType = "dev.tekton.events.plugin.step.succeeded.v1"
	CloudEventTypeStepSkipped   TektonPluginEventType = "dev.tekton.events.plugin.step.skipped.v1"
)

func (c TektonPluginEventType) String() string {
	return string(c)
}

type TektonStepCloudEvent struct {
	Pod    *corev1.Pod        `json:"pod,omitempty"`
	Logs   string             `json:"logs,omitempty"`
	State  *v1beta1.StepState `json:"state,omitempty"`
	Detail *v1beta1.Step      `json:"detail,omitempty"`
}

func NewTektonCloudEventStepCloudEvent(run *v1beta1.TaskRun, eventType TektonPluginEventType) TektonStepCloudEvent {
	return TektonStepCloudEvent{}
}

func EventForStep(run *v1beta1.TaskRun, eventType TektonPluginEventType) (*cloudevents.Event, error) {
	return nil, nil
}

func getPod() {}

func getLogs() {}
